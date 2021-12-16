package main

func getKarma(daysWithoutWeed int64, karma int64, smokedYesteday bool) (int64, int64) {
	f := int64(getFactor(daysWithoutWeed))

	if smokedYesteday {
		if daysWithoutWeed > 0 {
			daysWithoutWeed = 0
			karma /= f
		} else {
			daysWithoutWeed -= 1
			karma -= 10 * f
		}
	} else {
		if daysWithoutWeed < 0 {
			daysWithoutWeed = 1
		} else {
			daysWithoutWeed += 1
		}
		karma += 10 * f
	}

	return karma, daysWithoutWeed
}