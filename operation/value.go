package operation

type Value interface {
	Print(indent string) string
}
