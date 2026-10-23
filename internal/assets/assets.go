package assets

import _ "embed"

//go:embed favicon.png
var faviconPNG []byte

//go:embed Ringtone.mp3
var ringtoneMP3 []byte

//go:embed notification.wav
var notificationWAV []byte

func FaviconPNG() []byte { return faviconPNG }

func RingtoneMP3() []byte { return ringtoneMP3 }

func NotificationWAV() []byte { return notificationWAV }
