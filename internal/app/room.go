package app

func (a *App) findRoom(id string) (Room, bool) {
	for _, room := range a.cfg.Rooms {
		if room.ID == id {
			return room, true
		}
	}
	return Room{}, false
}
