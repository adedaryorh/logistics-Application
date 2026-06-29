package redis

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	platformconfig "github.com/adedaryorh/logistics-platform/pkg/config"
)

func TestMemoryStoreIncrWithTTL(t *testing.T) {
	store := NewMemoryStore()
	count, _, err := store.IncrWithTTL(context.Background(), "rate:key", time.Minute)
	if err != nil || count != 1 {
		t.Fatalf("expected first count 1, got count=%d err=%v", count, err)
	}
	count, _, _ = store.IncrWithTTL(context.Background(), "rate:key", time.Minute)
	if count != 2 {
		t.Fatalf("expected second count 2, got %d", count)
	}
}

func TestMemoryStoreSets(t *testing.T) {
	store := NewMemoryStore()
	if err := store.SAdd(context.Background(), "h3:1", "driver-1", "driver-2"); err != nil {
		t.Fatalf("SAdd() error = %v", err)
	}
	members, err := store.SMembers(context.Background(), "h3:1")
	if err != nil || len(members) != 2 {
		t.Fatalf("expected 2 members, got len=%d err=%v", len(members), err)
	}
}

func TestClientCommands(t *testing.T) {
	server := newFakeRedisServer(t)
	defer server.Close()

	client := NewClient(Options{Addr: server.Addr()})

	if err := client.Ping(context.Background()); err != nil {
		t.Fatalf("Ping() error = %v", err)
	}
	if err := client.Set(context.Background(), "greeting", "hello", time.Minute); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	got, ok, err := client.Get(context.Background(), "greeting")
	if err != nil || !ok || got != "hello" {
		t.Fatalf("Get() = %q, %v, %v", got, ok, err)
	}
	count, expiresAt, err := client.IncrWithTTL(context.Background(), "rate:key", time.Minute)
	if err != nil || count != 1 || expiresAt.Before(time.Now().UTC()) {
		t.Fatalf("IncrWithTTL() = %d, %v, %v", count, expiresAt, err)
	}
	if err := client.SAdd(context.Background(), "drivers", "driver-1", "driver-2"); err != nil {
		t.Fatalf("SAdd() error = %v", err)
	}
	members, err := client.SMembers(context.Background(), "drivers")
	if err != nil || len(members) != 2 {
		t.Fatalf("SMembers() len=%d err=%v", len(members), err)
	}
	if err := client.SRem(context.Background(), "drivers", "driver-1"); err != nil {
		t.Fatalf("SRem() error = %v", err)
	}
	members, err = client.SMembers(context.Background(), "drivers")
	if err != nil || len(members) != 1 || members[0] != "driver-2" {
		t.Fatalf("SMembers() after SRem = %+v err=%v", members, err)
	}
	if err := client.Delete(context.Background(), "greeting"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
}

func TestFallbackStoreFallsBackToMemory(t *testing.T) {
	primary := NewClient(Options{
		Addr:         "127.0.0.1:1",
		DialTimeout:  20 * time.Millisecond,
		ReadTimeout:  20 * time.Millisecond,
		WriteTimeout: 20 * time.Millisecond,
		MaxRetries:   0,
	})
	store := NewFallbackStore(primary, NewMemoryStore())

	if err := store.Set(context.Background(), "k", "v", time.Minute); err != nil {
		t.Fatalf("Set() fallback error = %v", err)
	}
	got, ok, err := store.Get(context.Background(), "k")
	if err != nil || !ok || got != "v" {
		t.Fatalf("Get() fallback = %q, %v, %v", got, ok, err)
	}
}

func TestNewStoreFromConfig(t *testing.T) {
	cfg := &platformconfig.Config{
		Redis: platformconfig.RedisConfig{
			Addr: "127.0.0.1:1",
			DB:   2,
		},
	}
	store := NewStoreFromConfig(cfg)
	if store == nil {
		t.Fatal("expected store from config")
	}
}

type fakeRedisServer struct {
	listener net.Listener
	mu       sync.Mutex
	values   map[string]string
	expires  map[string]time.Time
	sets     map[string]map[string]struct{}
}

func newFakeRedisServer(t *testing.T) *fakeRedisServer {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	server := &fakeRedisServer{
		listener: ln,
		values:   map[string]string{},
		expires:  map[string]time.Time{},
		sets:     map[string]map[string]struct{}{},
	}
	go server.serve(t)
	return server
}

func (s *fakeRedisServer) Addr() string { return s.listener.Addr().String() }

func (s *fakeRedisServer) Close() { _ = s.listener.Close() }

func (s *fakeRedisServer) serve(t *testing.T) {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return
		}
		go s.handleConn(t, conn)
	}
}

func (s *fakeRedisServer) handleConn(t *testing.T, conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	for {
		args, err := readRESPArray(reader)
		if err != nil {
			return
		}
		if len(args) == 0 {
			_, _ = writer.WriteString("-ERR empty command\r\n")
			_ = writer.Flush()
			continue
		}
		reply := s.exec(args)
		_, _ = writer.WriteString(reply)
		_ = writer.Flush()
	}
}

func (s *fakeRedisServer) exec(args []string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	for key, expiry := range s.expires {
		if now.After(expiry) {
			delete(s.values, key)
			delete(s.expires, key)
		}
	}

	switch strings.ToUpper(args[0]) {
	case "PING":
		return "+PONG\r\n"
	case "AUTH", "SELECT":
		return "+OK\r\n"
	case "SET":
		s.values[args[1]] = args[2]
		if len(args) >= 5 && strings.EqualFold(args[3], "PX") {
			ms, _ := strconv.Atoi(args[4])
			s.expires[args[1]] = now.Add(time.Duration(ms) * time.Millisecond)
		}
		return "+OK\r\n"
	case "GET":
		if value, ok := s.values[args[1]]; ok {
			return bulk(value)
		}
		return "$-1\r\n"
	case "DEL":
		delete(s.values, args[1])
		delete(s.expires, args[1])
		delete(s.sets, args[1])
		return ":1\r\n"
	case "INCR":
		current, _ := strconv.Atoi(s.values[args[1]])
		current++
		s.values[args[1]] = strconv.Itoa(current)
		return fmt.Sprintf(":%d\r\n", current)
	case "PEXPIRE":
		ms, _ := strconv.Atoi(args[2])
		s.expires[args[1]] = now.Add(time.Duration(ms) * time.Millisecond)
		return ":1\r\n"
	case "PTTL":
		expiry, ok := s.expires[args[1]]
		if !ok {
			return ":-1\r\n"
		}
		return fmt.Sprintf(":%d\r\n", int(expiry.Sub(now).Milliseconds()))
	case "SADD":
		if s.sets[args[1]] == nil {
			s.sets[args[1]] = map[string]struct{}{}
		}
		for _, member := range args[2:] {
			s.sets[args[1]][member] = struct{}{}
		}
		return fmt.Sprintf(":%d\r\n", len(args)-2)
	case "SREM":
		for _, member := range args[2:] {
			delete(s.sets[args[1]], member)
		}
		return fmt.Sprintf(":%d\r\n", len(args)-2)
	case "SMEMBERS":
		members := make([]string, 0, len(s.sets[args[1]]))
		for member := range s.sets[args[1]] {
			members = append(members, member)
		}
		var builder strings.Builder
		builder.WriteString(fmt.Sprintf("*%d\r\n", len(members)))
		for _, member := range members {
			builder.WriteString(bulk(member))
		}
		return builder.String()
	default:
		return "-ERR unsupported command\r\n"
	}
}

func readRESPArray(reader *bufio.Reader) ([]string, error) {
	header, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	if len(header) == 0 || header[0] != '*' {
		return nil, fmt.Errorf("expected array header")
	}
	count, err := strconv.Atoi(strings.TrimSpace(header[1:]))
	if err != nil {
		return nil, err
	}
	args := make([]string, 0, count)
	for i := 0; i < count; i++ {
		sizeLine, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		size, err := strconv.Atoi(strings.TrimSpace(sizeLine[1:]))
		if err != nil {
			return nil, err
		}
		buf := make([]byte, size+2)
		if _, err := ioReadAllFromBuf(reader, buf); err != nil {
			return nil, err
		}
		args = append(args, string(buf[:size]))
	}
	return args, nil
}

func ioReadAllFromBuf(reader *bufio.Reader, buf []byte) (int, error) {
	total := 0
	for total < len(buf) {
		n, err := reader.Read(buf[total:])
		total += n
		if err != nil {
			return total, err
		}
	}
	return total, nil
}

func bulk(value string) string {
	return fmt.Sprintf("$%d\r\n%s\r\n", len(value), value)
}
