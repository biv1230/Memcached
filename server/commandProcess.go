package server

import "fmt"

func ServiceCommand(c *Command) error {
	switch string(c.Name) {
	case CacheBeforeString:

	case CachingString:

	case CacheAfterString:

	default:
		return fmt.Errorf("unknow command %s", c.Name)
	}
	return nil
}
