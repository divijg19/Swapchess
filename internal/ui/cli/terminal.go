package cli

type Terminal interface {
	Size() (int, int, error)
	NextEvent() (KeyEvent, error)
	Render(Frame) error
	Close() error
}
