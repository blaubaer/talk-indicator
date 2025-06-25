package signal

type Signal interface {
	Dispose() error
	Ensure(Context) error
	Update() error

	GetType() Type
}
