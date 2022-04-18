package views

const (
	AlertLvlError   = "danger"
	AlertLvlWarning = "warning"
	AlertLvlInfo    = "info"
	AlertLvlSuccess = "success"
	//to be displayed when a random error is encountered by our backend
	AlertMagGeneric = "Something went wrong. Please try again, Contact support if the problem persists!"
)

//data in the top level structure that views expect data to come in
type Data struct {
	Alert *Alert
	Yield interface{}
}

//alert is used to render bootstrap alert messages in templates
type Alert struct {
	Level   string
	Message string
}

type PublicError interface {
	error
	Public() string
}

func (d *Data) AlertError(msg string) {
	d.Alert = &Alert{
		Level: AlertLvlError,
		Message: msg,
	}
}

//allows us to easily set an alert on our data type using an error
func (d *Data) SetAlert(err error) {
	var msg string
	if pErr, ok := err.(PublicError); ok {
		msg = pErr.Public()
	} else {
		msg = AlertMagGeneric
	}
	d.Alert = &Alert{
		Level:   AlertLvlError,
		Message: msg,
	}
}
