package internal

type ID interface {
	String() string
	Bytes() []byte
}

type IDer interface {
	GenID() (ID, error)
}
