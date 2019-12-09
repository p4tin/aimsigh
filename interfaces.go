package aimsigh

type Discoverer interface {
	UpdateAliveRecord(string, string, string, int) (RegistrationRecord, error)
	GetServiceAddress(string, string) (string, error)
}

type PersistenceDao interface {
}
