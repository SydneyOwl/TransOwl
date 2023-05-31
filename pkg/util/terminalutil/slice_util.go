package terminalutil

import "github.com/sydneyowl/TransOwl/internal/terminal"

func RemoveStringDuplicateUseMap(list []terminal.Terminal) []terminal.Terminal {
	var data []terminal.Terminal
	removeDupl := map[string]struct{}{}
	for _, v := range list {
		if _, ok := removeDupl[v.User.UserName]; !ok { //通过map内是否存在对应key值去添加对应切片内元素
			removeDupl[v.User.UserName] = struct{}{}
			data = append(data, v)
		}
	}
	return data
}
