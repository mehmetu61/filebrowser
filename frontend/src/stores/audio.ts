import { defineStore } from "pinia";

export interface AudioTrack {
  name: string;
  url: string;
  path: string;
}

export const useAudioStore = defineStore("audio", {
  state: (): {
    currentTrack: AudioTrack | null;
    isPlaying: boolean;
    playlist: AudioTrack[];
    currentIndex: number;
    currentTime: number;
    duration: number;
    volume: number;
  } => ({
    currentTrack: null,
    isPlaying: false,
    playlist: [],
    currentIndex: -1,
    currentTime: 0,
    duration: 0,
    volume: 1,
  }),

  actions: {
    playTrack(track: AudioTrack, playlist: AudioTrack[] = []) {
      this.currentTrack = track;
      this.playlist = playlist.length > 0 ? playlist : [track];
      this.currentIndex = this.playlist.findIndex((t) => t.path === track.path);
      this.isPlaying = true;
    },

    togglePlay() {
      this.isPlaying = !this.isPlaying;
    },

    next() {
      if (this.playlist.length === 0) return;
      this.currentIndex = (this.currentIndex + 1) % this.playlist.length;
      this.currentTrack = this.playlist[this.currentIndex];
      this.isPlaying = true;
    },

    prev() {
      if (this.playlist.length === 0) return;
      this.currentIndex =
        (this.currentIndex - 1 + this.playlist.length) % this.playlist.length;
      this.currentTrack = this.playlist[this.currentIndex];
      this.isPlaying = true;
    },

    stop() {
      this.currentTrack = null;
      this.isPlaying = false;
      this.currentTime = 0;
      this.duration = 0;
    },
  },
});
