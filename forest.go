package main

import (
	"context"
	"errors"

	"github.com/traPtitech/go-traq"
	traqwsbot "github.com/traPtitech/traq-ws-bot"
)

type Forest struct {
	channels   []traq.Channel
	path_to_id map[string]string
	id_to_path map[string]string
}

type resolvedChannel struct {
	Channel traq.Channel
	Path    string
}

func NewForest(bot *traqwsbot.Bot) (*Forest, error) {
	// get all channels
	channels, _, err := bot.API().ChannelApi.GetChannels(context.Background()).Execute()
	if err != nil {
		return nil, errors.New("traqforest.NewForest: failed to get channels")
	}
	// create id to channel map
	id_to_channel := make(map[string]traq.Channel)
	for _, channel := range channels.Public {
		id_to_channel[channel.Id] = channel
	}
	// create parent_id_to_id map
	parent_id_to_id := make(map[string][]string)
	roots_id := make([]string, 0)
	for _, channel := range channels.Public {
		if channel.ParentId.Get() != nil {
			parent_id := *(channel.ParentId.Get())
			channel_id := channel.Id
			if _, ok := parent_id_to_id[parent_id]; !ok {
				parent_id_to_id[parent_id] = make([]string, 0)
			}
			parent_id_to_id[parent_id] = append(parent_id_to_id[parent_id], channel_id)
		} else {
			roots_id = append(roots_id, channel.Id)
		}
	}
	// create path_to_id map
	// DFS forest from roots
	resolved_stack := make([]resolvedChannel, 0)
	resolved_vector := make([]resolvedChannel, 0)
	for _, root_id := range roots_id {
		resolved_stack = append(resolved_stack, resolvedChannel{
			Channel: id_to_channel[root_id],
			Path:    id_to_channel[root_id].Name,
		})
		resolved_vector = append(resolved_vector, resolvedChannel{
			Channel: id_to_channel[root_id],
			Path:    id_to_channel[root_id].Name,
		})
	}
	for len(resolved_stack) > 0 {
		// pop
		current := resolved_stack[len(resolved_stack)-1]
		resolved_stack = resolved_stack[:len(resolved_stack)-1]
		// push children
		if children, ok := parent_id_to_id[current.Channel.Id]; ok {
			for _, child_id := range children {
				resolved_stack = append(resolved_stack, resolvedChannel{
					Channel: id_to_channel[child_id],
					Path:    current.Path + "/" + id_to_channel[child_id].Name,
				})
				resolved_vector = append(resolved_vector, resolvedChannel{
					Channel: id_to_channel[child_id],
					Path:    current.Path + "/" + id_to_channel[child_id].Name,
				})
			}
		}
	}
	// create path_to_id and id_to_path map
	path_to_id := make(map[string]string)
	id_to_path := make(map[string]string)
	for _, resolved := range resolved_vector {
		path_to_id[resolved.Path] = resolved.Channel.Id
		id_to_path[resolved.Channel.Id] = resolved.Path
	}
	// return
	return &Forest{
		channels:   channels.Public,
		path_to_id: path_to_id,
		id_to_path: id_to_path,
	}, nil
}

func (f *Forest) GetChannel(path string) (traq.Channel, bool) {
	id, ok := f.path_to_id[path]
	if !ok {
		return traq.Channel{}, false
	}
	for _, channel := range f.channels {
		if channel.Id == id {
			return channel, true
		}
	}
	return traq.Channel{}, false
}

func (f *Forest) GetPath(id string) (string, bool) {
	path, ok := f.id_to_path[id]
	return path, ok
}
