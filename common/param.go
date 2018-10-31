package common

// ParamSetter for setting parameters
type ParamSetter interface {
	SetParam(name string, value string, kmsKey string) error
}

// ParamDeleter for deleting parameters
type ParamDeleter interface {
	DeleteParam(name string) error
}

// ParamGetter for getting parameters
type ParamGetter interface {
	GetParam(name string) (string, error)
}

// ParamManager composite of all param capabilities
type ParamManager interface {
	ParamGetter
	ParamSetter
	ParamDeleter
	ParamVersion(name string) (int64, error)
}
