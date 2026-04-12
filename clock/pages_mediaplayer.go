package clock

import (
	"image"
	"image/color"

	"github.com/g-wilson/led/internal/hamediaplayer"
	"github.com/g-wilson/led/internal/huegradient"
)

var mediaGradient = huegradient.Gradient{BaseHue: 160, Step: 110}

var nowPlayingTitle = color.RGBA{215, 0, 88, 255}
var nowPlayingMuted = color.RGBA{100, 100, 100, 255}

func (r *ClockRenderer) renderNowPlaying(c *image.RGBA) error {
	r.addText(c, image.Point{X: 0, Y: 5}, "Now Playing", nowPlayingTitle)

	if player, ok := r.mediaPlayer.GetPlayingPlayer(); ok {
		renderMediaInfo(r, c, player)
		return nil
	}

	r.addText(c, image.Point{X: 0, Y: 12}, "Nothing playing", nowPlayingMuted)
	return nil
}

func renderMediaInfo(r *ClockRenderer, c *image.RGBA, player hamediaplayer.MediaPlayerState) {
	if player.MediaArtist != "" {
		r.addText(c, image.Point{X: 0, Y: 12}, truncate16(player.MediaArtist), mediaGradient.Color(0))
	}
	if player.MediaTitle != "" {
		r.addText(c, image.Point{X: 0, Y: 18}, truncate16(player.MediaTitle), mediaGradient.Color(1))
	}
	if player.MediaAlbum != "" {
		r.addText(c, image.Point{X: 0, Y: 24}, truncate16(player.MediaAlbum), mediaGradient.Color(2))
	}
}

func truncate16(s string) string {
	r := []rune(s)
	if len(r) > 16 {
		return string(r[:16])
	}
	return s
}
