package stat

type EmptyTrafficMeter struct {
	TrafficMeter
}

func (t *EmptyTrafficMeter) Count(string, uint64, uint64) {
	//do nothing
}

func (t *EmptyTrafficMeter) Query(string) (uint64, uint64) {
	//do nothing
	return 0, 0
}

func (t *EmptyTrafficMeter) Close() error {
	//do nothing
	return nil
}

type EmptyAuthenticator struct {
	Authenticator
}

func (a *EmptyAuthenticator) CheckHash(hash string) bool {
	return true
}

func (a *EmptyAuthenticator) Close() error {
	return nil
}
