package rxfsnotify

type fileEvent struct {
	Path  string
	Event string
}

type CallBackEvent struct {
	Path  string
	Exist bool
}

type IPathCallback interface {
	OnPathChanged(cbe CallBackEvent)
}

var cb IPathCallback

func SetPathCallbackListener(_cb IPathCallback) {
	cb = _cb
}
