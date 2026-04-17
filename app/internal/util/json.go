package util

import "github.com/bytedance/sonic"

func MarshalJson(data any, format bool) ([]byte, error) {
	if format {
		return sonic.MarshalIndent(data, "", "\t")
	}

	b, err := sonic.Marshal(data)
	if err != nil {
		return nil, err
	}

	return b, nil
}
