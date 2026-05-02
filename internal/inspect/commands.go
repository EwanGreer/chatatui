package inspect

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/redis/go-redis/v9"
)

func (m Model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, scanKeysCmd(m.rdb), listChannelsCmd(m.rdb))
}

func scanKeysCmd(rdb *redis.Client) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var allKeys []string
		var cursor uint64
		for {
			keys, next, err := rdb.Scan(ctx, cursor, "*", 100).Result()
			if err != nil {
				return errMsg(err)
			}
			allKeys = append(allKeys, keys...)
			cursor = next
			if cursor == 0 {
				break
			}
		}
		sort.Strings(allKeys)
		return keysMsg(allKeys)
	}
}

func listChannelsCmd(rdb *redis.Client) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		channels, err := rdb.PubSubChannels(ctx, "*").Result()
		if err != nil {
			return errMsg(err)
		}
		sort.Strings(channels)
		return channelsMsg(channels)
	}
}

func subscribeChannel(rdb *redis.Client, channel string) (<-chan *redis.Message, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	sub := rdb.Subscribe(ctx, channel)
	ch := sub.Channel()

	return ch, func() {
		_ = sub.Close()
		cancel()
	}
}

func waitForChanMsg(ch <-chan *redis.Message) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return nil
		}
		return chanDataMsg(msg.Payload)
	}
}

func loadKeyDetailCmd(rdb *redis.Client, key string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		keyType, err := rdb.Type(ctx, key).Result()
		if err != nil {
			return errMsg(err)
		}

		ttl, err := rdb.TTL(ctx, key).Result()
		if err != nil {
			return errMsg(err)
		}

		value := fetchValue(ctx, rdb, key, keyType)

		return keyDetailMsg(KeyDetail{
			Key:   key,
			Type:  keyType,
			TTL:   ttl,
			Value: value,
		})
	}
}

func fetchValue(ctx context.Context, rdb *redis.Client, key, keyType string) string {
	const maxItems = 100

	switch keyType {
	case "string":
		val, err := rdb.Get(ctx, key).Result()
		if err != nil {
			return fmt.Sprintf("(error: %v)", err)
		}
		return formatStringValue(val)

	case "list":
		length, _ := rdb.LLen(ctx, key).Result()
		vals, err := rdb.LRange(ctx, key, 0, maxItems-1).Result()
		if err != nil {
			return fmt.Sprintf("(error: %v)", err)
		}
		var b strings.Builder
		for i, v := range vals {
			fmt.Fprintf(&b, "[%d] %s\n", i, v)
		}
		if length > maxItems {
			fmt.Fprintf(&b, "\n(%d of %d shown)", maxItems, length)
		}
		return b.String()

	case "set":
		vals, err := rdb.SMembers(ctx, key).Result()
		if err != nil {
			return fmt.Sprintf("(error: %v)", err)
		}
		sort.Strings(vals)
		if len(vals) > maxItems {
			vals = vals[:maxItems]
		}
		var b strings.Builder
		for _, v := range vals {
			fmt.Fprintf(&b, "• %s\n", v)
		}
		return b.String()

	case "zset":
		vals, err := rdb.ZRangeWithScores(ctx, key, 0, maxItems-1).Result()
		if err != nil {
			return fmt.Sprintf("(error: %v)", err)
		}
		var b strings.Builder
		for _, z := range vals {
			fmt.Fprintf(&b, "%-12g %v\n", z.Score, z.Member)
		}
		return b.String()

	case "hash":
		vals, err := rdb.HGetAll(ctx, key).Result()
		if err != nil {
			return fmt.Sprintf("(error: %v)", err)
		}
		keys := make([]string, 0, len(vals))
		for k := range vals {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var b strings.Builder
		maxFieldLen := 0
		for _, k := range keys {
			if len(k) > maxFieldLen {
				maxFieldLen = len(k)
			}
		}
		for _, k := range keys {
			fmt.Fprintf(&b, "%-*s  %s\n", maxFieldLen, k, vals[k])
		}
		return b.String()

	case "stream":
		msgs, err := rdb.XRevRange(ctx, key, "+", "-").Result()
		if err != nil {
			return fmt.Sprintf("(error: %v)", err)
		}
		if len(msgs) > 20 {
			msgs = msgs[:20]
		}
		var b strings.Builder
		for _, m := range msgs {
			fmt.Fprintf(&b, "%s\n", m.ID)
			for k, v := range m.Values {
				fmt.Fprintf(&b, "  %s: %v\n", k, v)
			}
		}
		return b.String()

	case "none":
		return "(key does not exist)"

	default:
		return fmt.Sprintf("(unsupported type: %s)", keyType)
	}
}

func formatStringValue(val string) string {
	if !utf8.ValidString(val) {
		return fmt.Sprintf("(binary data, %d bytes)", len(val))
	}

	trimmed := strings.TrimSpace(val)
	if len(trimmed) > 0 && (trimmed[0] == '{' || trimmed[0] == '[') {
		var buf strings.Builder
		if err := indentJSON(&buf, trimmed); err == nil {
			return buf.String()
		}
	}
	return val
}

func clearErrCmd() tea.Cmd {
	return tea.Tick(4*time.Second, func(time.Time) tea.Msg {
		return clearErrMsg{}
	})
}
