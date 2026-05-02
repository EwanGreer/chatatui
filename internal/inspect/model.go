package inspect

import (
	"context"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/redis/go-redis/v9"
)

type tab int

const (
	tabKeys tab = iota
	tabChannels
)

type focus int

const (
	focusSidebar focus = iota
	focusDetail
	focusFilter
)

type KeyDetail struct {
	Key   string
	Type  string
	TTL   time.Duration
	Value string
}

type Model struct {
	rdb      *redis.Client
	tab      tab
	keys     []string
	filtered []string
	filter   textinput.Model
	cursor   int
	detail   KeyDetail
	viewport viewport.Model
	width    int
	height   int
	ready    bool
	focus    focus
	loading  bool
	err      error

	channels       []string
	chanFiltered   []string
	chanCursor     int
	chanMessages   []string
	subscribedTo   string
	chanCancelFunc context.CancelFunc
	chanRecv       <-chan *redis.Message
}

type (
	keysMsg      []string
	keyDetailMsg KeyDetail
	channelsMsg  []string
	chanDataMsg  string
	errMsg       error
	clearErrMsg  struct{}
)

func NewModel(rdb *redis.Client) *Model {
	fi := textinput.New()
	fi.Placeholder = "filter..."
	fi.CharLimit = 256

	return &Model{
		rdb:    rdb,
		filter: fi,
		focus:  focusSidebar,
		tab:    tabKeys,
	}
}
