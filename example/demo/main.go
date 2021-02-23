package main

func main() {
	log.Std = false
	defer log.Close()

	if svr == nil {
		return
	}

	err := svr.Run()
	if err != nil {
		log.Error(err)
	}
}
