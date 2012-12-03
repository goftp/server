package raval

const (
	rootDir        = "/"
	welcomeMessage = "Welcome to the Go FTP Server"
	USER           = "USER"
	PASS           = "PASS"
)

type Array struct {
	container []interface{}
}

func (a *Array) Append(object interface{}) {
	if a.container == nil {
		a.container = make([]interface{}, 0)
	}
	newContainer := make([]interface{}, len(a.container)+1)
	copy(newContainer, a.container)
	newContainer[len(newContainer)-1] = object
	a.container = newContainer
}

func (a *Array) Remove(object interface{}) (result bool) {
	result = false
	newContainer := make([]interface{}, len(a.container)-1)
	i := 0
	for _, target := range a.container {
		if target != object {
			newContainer[i] = target
		} else {
			result = true
		}
		i++
	}
	return
}
