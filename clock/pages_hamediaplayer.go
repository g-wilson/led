package clock

import (
	"image"
	"image/color"

	"github.com/g-wilson/led/internal/hamediaplayer"
	"github.com/g-wilson/led/internal/huegradient"
)

var mediaGradient = huegradient.Gradient{BaseHue: 160, Step: 75}

var nowPlayingTitle = color.RGBA{215, 0, 88, 255}
var nowPlayingMuted = color.RGBA{100, 100, 100, 255}

func (r *ClockRenderer) renderNowPlaying(c *image.RGBA) error {
	player, ok := r.mediaPlayer.GetPlayingPlayer()

	if !ok {
		r.addText(c, image.Point{X: 0, Y: 12}, "Nothing playing", nowPlayingMuted)
		return nil
	}

	r.addText(c, image.Point{X: 0, Y: 5}, "> "+truncateN(player.FriendlyName, 14), nowPlayingTitle)
	renderMediaInfo(r, c, player)
	return nil
}

func renderMediaInfo(r *ClockRenderer, c *image.RGBA, player hamediaplayer.MediaPlayerState) {
	if player.MediaArtist != "" {
		r.addText(c, image.Point{X: 0, Y: 12}, truncateN(player.MediaArtist, 16), mediaGradient.Color(0))
	}
	if player.MediaTitle != "" {
		r.addText(c, image.Point{X: 0, Y: 18}, truncateN(player.MediaTitle, 16), mediaGradient.Color(1))
	}
	if player.MediaAlbum != "" {
		r.addText(c, image.Point{X: 0, Y: 24}, truncateN(player.MediaAlbum, 16), mediaGradient.Color(2))
	}
}

func truncateN(s string, n int) string {
	r := []rune(s)
	if len(r) > n {
		return string(r[:n])
	}
	return s
}
