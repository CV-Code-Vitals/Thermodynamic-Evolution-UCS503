package reliability

import "os"

func SafeErrorCheck1(p string) error {
	if err := os.Remove(p); err != nil {
		return err
	}
	return nil
}
