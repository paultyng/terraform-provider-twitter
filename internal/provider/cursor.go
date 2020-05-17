package provider

func allPages(page func(cursor int64) (int64, error)) error {
	var err error
	var next int64 = -1
	for {
		next, err = page(next)
		if err != nil {
			return err
		}
		if next == 0 {
			return nil
		}
	}
}
