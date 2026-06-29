package tracking

type KalmanFilter struct {
	state            [4]float64
	covariance       [4][4]float64
	processNoise     float64
	measurementNoise float64
}

func NewKalmanFilter(initialLat, initialLng float64) *KalmanFilter {
	kf := &KalmanFilter{
		state:            [4]float64{initialLat, initialLng, 0, 0},
		processNoise:     0.0001,
		measurementNoise: 0.01,
	}
	for i := 0; i < 4; i++ {
		kf.covariance[i][i] = 1
	}
	return kf
}

func (kf *KalmanFilter) Update(lat, lng float64, dt float64) (smoothedLat, smoothedLng float64) {
	if dt <= 0 {
		dt = 1
	}

	kf.state[0] += kf.state[2] * dt
	kf.state[1] += kf.state[3] * dt

	kf.state[2] = 0.8*kf.state[2] + 0.2*((lat-kf.state[0])/dt)
	kf.state[3] = 0.8*kf.state[3] + 0.2*((lng-kf.state[1])/dt)

	blend := kf.processNoise / (kf.processNoise + kf.measurementNoise)
	kf.state[0] = (1-blend)*kf.state[0] + blend*lat
	kf.state[1] = (1-blend)*kf.state[1] + blend*lng

	return kf.state[0], kf.state[1]
}
