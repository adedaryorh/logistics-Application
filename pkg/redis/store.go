package redis

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/adedaryorh/logistics-platform/pkg/observability"
)

type Store interface {
	Ping(ctx context.Context) error
	IncrWithTTL(ctx context.Context, key string, ttl time.Duration) (int, time.Time, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Get(ctx context.Context, key string) (string, bool, error)
	Delete(ctx context.Context, key string) error
	SAdd(ctx context.Context, key string, members ...string) error
	SRem(ctx context.Context, key string, members ...string) error
	SMembers(ctx context.Context, key string) ([]string, error)
}

type Options struct {
	Addr         string
	Password     string
	DB           int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	MaxRetries   int
	RetryBackoff time.Duration
}

type Client struct {
	opts Options
}

type FallbackStore struct {
	primary  Store
	fallback Store
}

type MemoryStore struct {
	mu       sync.Mutex
	values   map[string]valueEntry
	sets     map[string]map[string]struct{}
	counters map[string]counterEntry
}

type valueEntry struct {
	value     string
	expiresAt time.Time
}

type counterEntry struct {
	count     int
	expiresAt time.Time
}

func DefaultOptions() Options {
	return Options{
		Addr:         "localhost:6379",
		DB:           0,
		DialTimeout:  2 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
		MaxRetries:   2,
		RetryBackoff: 100 * time.Millisecond,
	}
}

func NewClient(opts Options) *Client {
	defaults := DefaultOptions()
	if strings.TrimSpace(opts.Addr) == "" {
		opts.Addr = defaults.Addr
	}
	if opts.DialTimeout <= 0 {
		opts.DialTimeout = defaults.DialTimeout
	}
	if opts.ReadTimeout <= 0 {
		opts.ReadTimeout = defaults.ReadTimeout
	}
	if opts.WriteTimeout <= 0 {
		opts.WriteTimeout = defaults.WriteTimeout
	}
	if opts.MaxRetries < 0 {
		opts.MaxRetries = defaults.MaxRetries
	}
	if opts.RetryBackoff <= 0 {
		opts.RetryBackoff = defaults.RetryBackoff
	}
	return &Client{opts: opts}
}

func NewFallbackStore(primary, fallback Store) *FallbackStore {
	if fallback == nil {
		fallback = NewMemoryStore()
	}
	return &FallbackStore{primary: primary, fallback: fallback}
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		values:   map[string]valueEntry{},
		sets:     map[string]map[string]struct{}{},
		counters: map[string]counterEntry{},
	}
}

func (s *FallbackStore) Ping(ctx context.Context) error {
	if s.primary != nil {
		if err := s.primary.Ping(ctx); err == nil {
			observability.IncCounter("redis_fallback_requests_total", 1, map[string]string{"command": "PING", "backend": "primary", "status": "success"})
			return nil
		}
	}
	if s.fallback != nil {
		err := s.fallback.Ping(ctx)
		status := "success"
		if err != nil {
			status = "failed"
		}
		observability.IncCounter("redis_fallback_requests_total", 1, map[string]string{"command": "PING", "backend": "fallback", "status": status})
		return err
	}
	return nil
}

func (s *FallbackStore) IncrWithTTL(ctx context.Context, key string, ttl time.Duration) (int, time.Time, error) {
	if s.primary != nil {
		count, expiresAt, err := s.primary.IncrWithTTL(ctx, key, ttl)
		if err == nil {
			observability.IncCounter("redis_fallback_requests_total", 1, map[string]string{"command": "INCR", "backend": "primary", "status": "success"})
			return count, expiresAt, nil
		}
	}
	observability.IncCounter("redis_fallback_requests_total", 1, map[string]string{"command": "INCR", "backend": "fallback", "status": "success"})
	return s.fallback.IncrWithTTL(ctx, key, ttl)
}

func (s *FallbackStore) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	if s.primary != nil {
		if err := s.primary.Set(ctx, key, value, ttl); err == nil {
			observability.IncCounter("redis_fallback_requests_total", 1, map[string]string{"command": "SET", "backend": "primary", "status": "success"})
			return nil
		}
	}
	observability.IncCounter("redis_fallback_requests_total", 1, map[string]string{"command": "SET", "backend": "fallback", "status": "success"})
	return s.fallback.Set(ctx, key, value, ttl)
}

func (s *FallbackStore) Get(ctx context.Context, key string) (string, bool, error) {
	if s.primary != nil {
		value, ok, err := s.primary.Get(ctx, key)
		if err == nil {
			status := "hit"
			if !ok {
				status = "miss"
			}
			observability.IncCounter("redis_fallback_requests_total", 1, map[string]string{"command": "GET", "backend": "primary", "status": status})
			return value, ok, nil
		}
	}
	observability.IncCounter("redis_fallback_requests_total", 1, map[string]string{"command": "GET", "backend": "fallback", "status": "success"})
	return s.fallback.Get(ctx, key)
}

func (s *FallbackStore) Delete(ctx context.Context, key string) error {
	if s.primary != nil {
		if err := s.primary.Delete(ctx, key); err == nil {
			observability.IncCounter("redis_fallback_requests_total", 1, map[string]string{"command": "DEL", "backend": "primary", "status": "success"})
			return nil
		}
	}
	observability.IncCounter("redis_fallback_requests_total", 1, map[string]string{"command": "DEL", "backend": "fallback", "status": "success"})
	return s.fallback.Delete(ctx, key)
}

func (s *FallbackStore) SAdd(ctx context.Context, key string, members ...string) error {
	if s.primary != nil {
		if err := s.primary.SAdd(ctx, key, members...); err == nil {
			observability.IncCounter("redis_fallback_requests_total", 1, map[string]string{"command": "SADD", "backend": "primary", "status": "success"})
			return nil
		}
	}
	observability.IncCounter("redis_fallback_requests_total", 1, map[string]string{"command": "SADD", "backend": "fallback", "status": "success"})
	return s.fallback.SAdd(ctx, key, members...)
}

func (s *FallbackStore) SRem(ctx context.Context, key string, members ...string) error {
	if s.primary != nil {
		if err := s.primary.SRem(ctx, key, members...); err == nil {
			observability.IncCounter("redis_fallback_requests_total", 1, map[string]string{"command": "SREM", "backend": "primary", "status": "success"})
			return nil
		}
	}
	observability.IncCounter("redis_fallback_requests_total", 1, map[string]string{"command": "SREM", "backend": "fallback", "status": "success"})
	return s.fallback.SRem(ctx, key, members...)
}

func (s *FallbackStore) SMembers(ctx context.Context, key string) ([]string, error) {
	if s.primary != nil {
		members, err := s.primary.SMembers(ctx, key)
		if err == nil {
			observability.IncCounter("redis_fallback_requests_total", 1, map[string]string{"command": "SMEMBERS", "backend": "primary", "status": "success"})
			return members, nil
		}
	}
	observability.IncCounter("redis_fallback_requests_total", 1, map[string]string{"command": "SMEMBERS", "backend": "fallback", "status": "success"})
	return s.fallback.SMembers(ctx, key)
}

func (c *Client) Ping(ctx context.Context) error {
	_, err := c.do(ctx, "PING")
	return err
}

func (c *Client) IncrWithTTL(ctx context.Context, key string, ttl time.Duration) (int, time.Time, error) {
	reply, err := c.do(ctx, "INCR", key)
	if err != nil {
		return 0, time.Time{}, err
	}
	count, ok := reply.(int64)
	if !ok {
		return 0, time.Time{}, fmt.Errorf("unexpected INCR reply type %T", reply)
	}

	if count == 1 {
		if _, err := c.do(ctx, "PEXPIRE", key, strconv.FormatInt(ttl.Milliseconds(), 10)); err != nil {
			return 0, time.Time{}, err
		}
		return int(count), time.Now().UTC().Add(ttl), nil
	}

	pttlReply, err := c.do(ctx, "PTTL", key)
	if err != nil {
		return 0, time.Time{}, err
	}
	pttl, ok := pttlReply.(int64)
	if !ok {
		return 0, time.Time{}, fmt.Errorf("unexpected PTTL reply type %T", pttlReply)
	}
	expiresAt := time.Now().UTC()
	if pttl > 0 {
		expiresAt = expiresAt.Add(time.Duration(pttl) * time.Millisecond)
	}
	return int(count), expiresAt, nil
}

func (c *Client) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	args := []string{"SET", key, value}
	if ttl > 0 {
		args = append(args, "PX", strconv.FormatInt(ttl.Milliseconds(), 10))
	}
	_, err := c.do(ctx, args...)
	return err
}

func (c *Client) Get(ctx context.Context, key string) (string, bool, error) {
	reply, err := c.do(ctx, "GET", key)
	if err != nil {
		return "", false, err
	}
	if reply == nil {
		return "", false, nil
	}
	value, ok := reply.(string)
	if !ok {
		return "", false, fmt.Errorf("unexpected GET reply type %T", reply)
	}
	return value, true, nil
}

func (c *Client) Delete(ctx context.Context, key string) error {
	_, err := c.do(ctx, "DEL", key)
	return err
}

func (c *Client) SAdd(ctx context.Context, key string, members ...string) error {
	args := append([]string{"SADD", key}, members...)
	_, err := c.do(ctx, args...)
	return err
}

func (c *Client) SRem(ctx context.Context, key string, members ...string) error {
	args := append([]string{"SREM", key}, members...)
	_, err := c.do(ctx, args...)
	return err
}

func (c *Client) SMembers(ctx context.Context, key string) ([]string, error) {
	reply, err := c.do(ctx, "SMEMBERS", key)
	if err != nil {
		return nil, err
	}
	if reply == nil {
		return []string{}, nil
	}
	items, ok := reply.([]any)
	if !ok {
		return nil, fmt.Errorf("unexpected SMEMBERS reply type %T", reply)
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, fmt.Sprintf("%v", item))
	}
	return out, nil
}

func (c *Client) do(ctx context.Context, args ...string) (any, error) {
	command := "unknown"
	if len(args) > 0 {
		command = strings.ToUpper(args[0])
	}
	var lastErr error
	for attempt := 0; attempt <= c.opts.MaxRetries; attempt++ {
		start := time.Now()
		reply, err := c.doOnce(ctx, args...)
		observability.ObserveHistogram("redis_client_duration_seconds", time.Since(start).Seconds(), map[string]string{
			"command": command,
			"attempt": strconv.Itoa(attempt + 1),
		})
		if err == nil {
			observability.IncCounter("redis_client_requests_total", 1, map[string]string{"command": command, "status": "success"})
			return reply, nil
		}
		lastErr = err
		observability.IncCounter("redis_client_requests_total", 1, map[string]string{"command": command, "status": "failed"})
		if attempt == c.opts.MaxRetries || ctx.Err() != nil {
			break
		}
		observability.IncCounter("redis_client_retries_total", 1, map[string]string{"command": command})
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(c.opts.RetryBackoff * time.Duration(attempt+1)):
		}
	}
	return nil, lastErr
}

func (c *Client) doOnce(ctx context.Context, args ...string) (any, error) {
	dialer := &net.Dialer{Timeout: c.opts.DialTimeout}
	conn, err := dialer.DialContext(ctx, "tcp", c.opts.Addr)
	if err != nil {
		return nil, fmt.Errorf("dial redis: %w", err)
	}
	defer conn.Close()

	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	} else {
		_ = conn.SetReadDeadline(time.Now().Add(c.opts.ReadTimeout))
		_ = conn.SetWriteDeadline(time.Now().Add(c.opts.WriteTimeout))
	}

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	if c.opts.Password != "" {
		if err := writeCommand(writer, "AUTH", c.opts.Password); err != nil {
			return nil, err
		}
		if err := writer.Flush(); err != nil {
			return nil, fmt.Errorf("flush redis auth: %w", err)
		}
		if _, err := readReply(reader); err != nil {
			return nil, err
		}
	}
	if c.opts.DB > 0 {
		if err := writeCommand(writer, "SELECT", strconv.Itoa(c.opts.DB)); err != nil {
			return nil, err
		}
		if err := writer.Flush(); err != nil {
			return nil, fmt.Errorf("flush redis select: %w", err)
		}
		if _, err := readReply(reader); err != nil {
			return nil, err
		}
	}

	if err := writeCommand(writer, args...); err != nil {
		return nil, err
	}
	if err := writer.Flush(); err != nil {
		return nil, fmt.Errorf("flush redis command: %w", err)
	}
	return readReply(reader)
}

func writeCommand(writer *bufio.Writer, args ...string) error {
	if _, err := fmt.Fprintf(writer, "*%d\r\n", len(args)); err != nil {
		return fmt.Errorf("write redis array header: %w", err)
	}
	for _, arg := range args {
		if _, err := fmt.Fprintf(writer, "$%d\r\n%s\r\n", len(arg), arg); err != nil {
			return fmt.Errorf("write redis bulk arg: %w", err)
		}
	}
	return nil
}

func readReply(reader *bufio.Reader) (any, error) {
	prefix, err := reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("read redis reply type: %w", err)
	}
	switch prefix {
	case '+':
		line, err := readLine(reader)
		if err != nil {
			return nil, err
		}
		return line, nil
	case '-':
		line, err := readLine(reader)
		if err != nil {
			return nil, err
		}
		return nil, errors.New(line)
	case ':':
		line, err := readLine(reader)
		if err != nil {
			return nil, err
		}
		value, err := strconv.ParseInt(line, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse redis integer: %w", err)
		}
		return value, nil
	case '$':
		line, err := readLine(reader)
		if err != nil {
			return nil, err
		}
		size, err := strconv.Atoi(line)
		if err != nil {
			return nil, fmt.Errorf("parse redis bulk size: %w", err)
		}
		if size == -1 {
			return nil, nil
		}
		buf := make([]byte, size+2)
		if _, err := ioReadFull(reader, buf); err != nil {
			return nil, fmt.Errorf("read redis bulk body: %w", err)
		}
		return string(buf[:size]), nil
	case '*':
		line, err := readLine(reader)
		if err != nil {
			return nil, err
		}
		size, err := strconv.Atoi(line)
		if err != nil {
			return nil, fmt.Errorf("parse redis array size: %w", err)
		}
		if size == -1 {
			return nil, nil
		}
		items := make([]any, 0, size)
		for i := 0; i < size; i++ {
			item, err := readReply(reader)
			if err != nil {
				return nil, err
			}
			items = append(items, item)
		}
		return items, nil
	default:
		return nil, fmt.Errorf("unsupported redis reply type %q", string(prefix))
	}
}

func readLine(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("read redis line: %w", err)
	}
	return strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r"), nil
}

func ioReadFull(reader *bufio.Reader, buf []byte) (int, error) {
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

func (s *MemoryStore) Ping(ctx context.Context) error {
	return nil
}

func (s *MemoryStore) IncrWithTTL(ctx context.Context, key string, ttl time.Duration) (int, time.Time, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	entry, ok := s.counters[key]
	if !ok || now.After(entry.expiresAt) {
		entry = counterEntry{count: 0, expiresAt: now.Add(ttl)}
	}
	entry.count++
	s.counters[key] = entry
	return entry.count, entry.expiresAt, nil
}

func (s *MemoryStore) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	expiresAt := time.Time{}
	if ttl > 0 {
		expiresAt = time.Now().UTC().Add(ttl)
	}
	s.values[key] = valueEntry{value: value, expiresAt: expiresAt}
	return nil
}

func (s *MemoryStore) Get(ctx context.Context, key string) (string, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.values[key]
	if !ok {
		return "", false, nil
	}
	if !entry.expiresAt.IsZero() && time.Now().UTC().After(entry.expiresAt) {
		delete(s.values, key)
		return "", false, nil
	}
	return entry.value, true, nil
}

func (s *MemoryStore) Delete(ctx context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.values, key)
	delete(s.sets, key)
	delete(s.counters, key)
	return nil
}

func (s *MemoryStore) SAdd(ctx context.Context, key string, members ...string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.sets[key] == nil {
		s.sets[key] = map[string]struct{}{}
	}
	for _, member := range members {
		s.sets[key][member] = struct{}{}
	}
	return nil
}

func (s *MemoryStore) SRem(ctx context.Context, key string, members ...string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, member := range members {
		delete(s.sets[key], member)
	}
	return nil
}

func (s *MemoryStore) SMembers(ctx context.Context, key string) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]string, 0, len(s.sets[key]))
	for member := range s.sets[key] {
		result = append(result, member)
	}
	return result, nil
}
