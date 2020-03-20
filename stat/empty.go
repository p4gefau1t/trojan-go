package stat

type EmptyTrafficMeter struct {
	TrafficMeter
}

func (t *EmptyTrafficMeter) Count(string, int, int) {
	//do nothing
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
