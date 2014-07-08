package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

const (
	gXiamiSongSplitter = ","
	gXiamiRetOK        = "ok"
	gXiamiRetFail      = "failed"
	gXiamiProvider     = "xiami"
	gXiamiAPIUrlBase   = "http://www.xiami.com/app"
	gXiamiSongUrl      = "/android/song/id/"
	gXiamiAlbumUrl     = "/iphone/album/id/"
	gXiamiCollectUrl   = "/android/collect?id="
)

var (
	gXiamiClient = &http.Client{}
)

type XiamiRetStatus struct {
	Status  string `json:"status"`
	Message string `json:"msg"`
}

type XiamiSongRet struct {
	XiamiRetStatus
	Song XiamiSong `json:"song"`
}

type XiamiSong struct {
	Name   string `json:"song_name"`
	Url    string `json:"song_location"`
	Lrc    string `json:"song_lrc"`
	Artist string `json:"artist_name"`
}

type XiamiCollectRet struct {
	XiamiRetStatus
	Collect XiamiCollect `json:"collect"`
}

type XiamiCollect struct {
	Songs []XiamiCollectSong `json:"songs"`
}

type XiamiCollectSong struct {
	Name   string `json:"name"`
	Url    string `json:"location"`
	Lrc    string `json:"lyric"`
	Artist string `json:"singers"`
}

type XiamiAlbumRet struct {
	XiamiRetStatus
	Album XiamiAlbum `json:"album"`
}

type XiamiAlbum struct {
	Songs map[string]XiamiCollectSong
}

func GetXiamiSong(songId string) *SongList {
	sl := NewSongList()
	ret := GetCache(gXiamiProvider, gReqTypeSongList, songId)
	url := gXiamiAPIUrlBase + gXiamiSongUrl + strings.TrimSpace(songId)
	if nil == ret {
		ret = GetUrl(gXiamiClient, url)
		if nil == ret {
			return sl.SetAndLogErrorf("error accessing url %s", url)
		}
		SetCache(gXiamiProvider, gReqTypeSong, songId, time.Duration(*gFlagCacheExpiration)*time.Second, ret)
	}

	var songret XiamiSongRet
	err := json.Unmarshal(ret, &songret)
	if nil != err {
		return sl.SetAndLogErrorf("error parsing song info from url %s: %s", url, err)
	}
	if gXiamiRetOK != songret.Status {
		return sl.SetAndLogErrorf("error getting url %s: %s", url, songret.Message)
	}
	emptyXiamiSong := XiamiSong{}
	if emptyXiamiSong == songret.Song {
		return sl.SetAndLogErrorf("invalid song id %s", songId)
	}
	return sl.AddSong(&Song{
		Name:     songret.Song.Name,
		Url:      songret.Song.Url,
		Artists:  songret.Song.Artist,
		LrcUrl:   songret.Song.Lrc,
		Provider: gXiamiProvider,
	})
}

func GetXiamiSongList(songs string) *SongList {
	sl := NewSongList()
	for _, sid := range strings.Split(songs, gXiamiSongSplitter) {
		singleSL := GetXiamiSong(strings.TrimSpace(sid))
		if singleSL.IsFailed() {
			return singleSL
		}
		sl.Concat(singleSL)
	}
	return sl
}

func GetXiamiCollect(collectId string) *SongList {
	url := gXiamiAPIUrlBase + gXiamiCollectUrl + strings.TrimSpace(collectId)
	ret := GetUrl(gXiamiClient, url)
	sl := NewSongList()
	if nil == ret {
		return sl.SetAndLogErrorf("error accessing url %s", url)
	}
	var collectRet XiamiCollectRet
	err := json.Unmarshal(ret, &collectRet)
	if nil != err {
		return sl.SetAndLogErrorf("error parsing collect data from url %s: %s", url, err)
	}
	if gXiamiRetOK != collectRet.Status {
		return sl.SetAndLogErrorf("error getting url %s: %s", url, collectRet.Message)
	}
	for i, _ := range collectRet.Collect.Songs {
		song := &collectRet.Collect.Songs[i]
		sl.AddSong(&Song{
			Name:     song.Name,
			Url:      song.Url,
			LrcUrl:   song.Lrc,
			Artists:  song.Artist,
			Provider: gXiamiProvider,
		})
	}
	return sl
}

func GetXiamiAlbum(albumId string) *SongList {
	url := gXiamiAPIUrlBase + gXiamiAlbumUrl + strings.TrimSpace(albumId)
	ret := GetUrl(gXiamiClient, url)
	sl := NewSongList()
	if nil == ret {
		return sl.SetAndLogErrorf("error accessing url %s", url)
	}
	var albumRet XiamiAlbumRet
	err := json.Unmarshal(ret, &albumRet)
	if nil != err {
		return sl.SetAndLogErrorf("error parsing album data from url %s: %s", url, err)
	}
	if gXiamiRetOK != albumRet.Status {
		return sl.SetAndLogErrorf("error getting url %s: %s", url, albumRet.Message)
	}
	for k, _ := range albumRet.Album.Songs {
		song := albumRet.Album.Songs[k]
		sl.AddSong(&Song{
			Name:     song.Name,
			Url:      song.Url,
			LrcUrl:   song.Lrc,
			Artists:  song.Artist,
			Provider: gXiamiProvider,
		})
	}
	return sl
}
