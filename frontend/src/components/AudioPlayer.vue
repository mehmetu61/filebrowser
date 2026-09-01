<template>
  <div v-if="audioStore.currentTrack" class="global-audio-player">
    <audio
      ref="audioRef"
      :src="audioStore.currentTrack.url"
      @timeupdate="onTimeUpdate"
      @loadedmetadata="onLoadedMetadata"
      @ended="onEnded"
    ></audio>

    <div class="audio-info">
      <i class="material-icons audio-icon">audiotrack</i>
      <div class="audio-title" :title="audioStore.currentTrack.name">
        {{ audioStore.currentTrack.name }}
      </div>
    </div>

    <div class="audio-controls">
      <button class="audio-btn" @click="audioStore.prev()" title="Previous track">
        <i class="material-icons">skip_previous</i>
      </button>

      <button class="audio-btn play-btn" @click="togglePlay" title="Play/Pause">
        <i class="material-icons">{{ audioStore.isPlaying ? "pause" : "play_arrow" }}</i>
      </button>

      <button class="audio-btn" @click="audioStore.next()" title="Next track">
        <i class="material-icons">skip_next</i>
      </button>

      <span class="audio-time">{{ formatTime(audioStore.currentTime) }} / {{ formatTime(audioStore.duration) }}</span>

      <input
        type="range"
        class="audio-progress"
        min="0"
        :max="audioStore.duration || 100"
        :value="audioStore.currentTime"
        @input="onSeek"
      />
    </div>

    <div class="audio-actions">
      <button class="audio-btn close-btn" @click="audioStore.stop()" title="Close player">
        <i class="material-icons">close</i>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from "vue";
import { useAudioStore } from "@/stores/audio";

const audioStore = useAudioStore();
const audioRef = ref<HTMLAudioElement | null>(null);

watch(
  () => audioStore.currentTrack,
  () => {
    if (audioRef.value && audioStore.isPlaying) {
      audioRef.value.play().catch(() => {});
    }
  }
);

watch(
  () => audioStore.isPlaying,
  (playing) => {
    if (!audioRef.value) return;
    if (playing) {
      audioRef.value.play().catch(() => {});
    } else {
      audioRef.value.pause();
    }
  }
);

const togglePlay = () => {
  audioStore.togglePlay();
};

const onTimeUpdate = () => {
  if (audioRef.value) {
    audioStore.currentTime = audioRef.value.currentTime;
  }
};

const onLoadedMetadata = () => {
  if (audioRef.value) {
    audioStore.duration = audioRef.value.duration;
    if (audioStore.isPlaying) {
      audioRef.value.play().catch(() => {});
    }
  }
};

const onEnded = () => {
  audioStore.next();
};

const onSeek = (e: Event) => {
  const val = Number((e.target as HTMLInputElement).value);
  if (audioRef.value) {
    audioRef.value.currentTime = val;
    audioStore.currentTime = val;
  }
};

const formatTime = (seconds: number): string => {
  if (!seconds || isNaN(seconds)) return "0:00";
  const mins = Math.floor(seconds / 60);
  const secs = Math.floor(seconds % 60);
  return `${mins}:${secs < 10 ? "0" : ""}${secs}`;
};
</script>

<style scoped>
.global-audio-player {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  height: 64px;
  background: var(--card-bg, #ffffff);
  color: var(--text-color, #2c3e50);
  border-top: 1px solid rgba(128, 128, 128, 0.2);
  box-shadow: 0 -4px 16px rgba(0, 0, 0, 0.15);
  display: flex;
  align-items: center;
  padding: 0 1.5em;
  gap: 1.5em;
  z-index: 1000;
  backdrop-filter: blur(8px);
}

.audio-info {
  display: flex;
  align-items: center;
  gap: 10px;
  max-width: 250px;
  min-width: 150px;
}

.audio-icon {
  color: #2196f3;
  font-size: 24px;
}

.audio-title {
  font-size: 14px;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.audio-controls {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 12px;
  justify-content: center;
}

.audio-btn {
  background: transparent;
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: inherit;
  opacity: 0.85;
  transition: opacity 0.15s;
  padding: 4px;
  border-radius: 50%;
}

.audio-btn:hover {
  opacity: 1;
  background: rgba(128, 128, 128, 0.15);
}

.play-btn {
  background: #2196f3;
  color: #ffffff;
  width: 38px;
  height: 38px;
}

.play-btn:hover {
  background: #1976d2;
}

.audio-time {
  font-size: 12px;
  font-family: monospace;
  opacity: 0.75;
  white-space: nowrap;
}

.audio-progress {
  flex: 1;
  max-width: 400px;
  cursor: pointer;
  accent-color: #2196f3;
}

.audio-actions {
  display: flex;
  align-items: center;
}

.close-btn {
  opacity: 0.6;
}

.close-btn:hover {
  opacity: 1;
}
</style>
