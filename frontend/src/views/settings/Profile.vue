<template>
  <div class="row">
    <div class="column">
      <form class="card" @submit="updateSettings">
        <div class="card-title">
          <h2>{{ t("settings.profileSettings") }}</h2>
        </div>

        <div class="card-content">
          <p>
            <input type="checkbox" name="hideDotfiles" v-model="hideDotfiles" />
            {{ t("settings.hideDotfiles") }}
          </p>
          <p>
            <input type="checkbox" name="singleClick" v-model="singleClick" />
            {{ t("settings.singleClick") }}
          </p>
          <p>
            <input
              type="checkbox"
              name="redirectAfterCopyMove"
              v-model="redirectAfterCopyMove"
            />
            {{ t("settings.redirectAfterCopyMove") }}
          </p>
          <p>
            <input type="checkbox" name="dateFormat" v-model="dateFormat" />
            {{ t("settings.setDateFormat") }}
          </p>
          <h3>{{ t("settings.language") }}</h3>
          <languages
            class="input input--block"
            v-model:locale="locale"
          ></languages>

          <h3>{{ t("settings.aceEditorTheme") }}</h3>
          <AceEditorTheme
            class="input input--block"
            v-model:aceEditorTheme="aceEditorTheme"
            id="aceTheme"
          ></AceEditorTheme>

          <h3>Theme</h3>
          <select
            class="input input--block"
            v-model="selectedTheme"
            @change="layoutStore.setTheme(selectedTheme)"
          >
            <option value="system">🌓 System Default</option>
            <option value="light">☀️ Light Theme</option>
            <option value="dark">🌙 Dark Slate Theme</option>
            <option value="oled">🖤 OLED Pure Black</option>
          </select>
        </div>

        <div class="card-action">
          <input
            class="button button--flat"
            type="submit"
            name="submitProfile"
            :value="t('buttons.update')"
          />
        </div>
      </form>
    </div>

    <div v-if="!noAuth" class="column">
      <form
        class="card"
        v-if="!authStore.user?.lockPassword"
        @submit="updatePassword"
      >
        <div class="card-title">
          <h2>{{ t("settings.changePassword") }}</h2>
        </div>

        <div class="card-content">
          <input
            :class="passwordClass"
            type="password"
            :placeholder="t('settings.newPassword')"
            v-model="password"
            name="password"
          />
          <input
            :class="passwordClass"
            type="password"
            :placeholder="t('settings.newPasswordConfirm')"
            v-model="passwordConf"
            name="passwordConf"
          />
          <input
            v-if="isCurrentPasswordRequired"
            :class="passwordClass"
            type="password"
            :placeholder="t('settings.currentPassword')"
            v-model="currentPassword"
            name="current_password"
            autocomplete="current-password"
          />
        </div>

        <div class="card-action">
          <input
            class="button button--flat"
            type="submit"
            name="submitPassword"
            :value="t('buttons.update')"
          />
        </div>
      </form>

      <!-- 2FA Configuration Card -->
      <div class="card" style="margin-top: 1.5em;">
        <div class="card-title">
          <h2>🔒 Two-Factor Authentication (2FA / TOTP)</h2>
        </div>

        <div class="card-content">
          <div v-if="authStore.user?.totpEnabled">
            <p style="color: #27ae60; font-weight: bold; margin-bottom: 1em;">
              ✓ Two-factor authentication is active on your account.
            </p>
            <div v-if="showDisable2FA">
              <input
                class="input input--block"
                type="password"
                placeholder="Confirm password to disable 2FA"
                v-model="disable2FAPassword"
              />
              <button
                class="button button--red"
                style="margin-top: 0.5em;"
                @click="onDisable2FA"
              >
                Disable 2FA
              </button>
              <button
                class="button button--flat"
                style="margin-top: 0.5em; margin-left: 0.5em;"
                @click="showDisable2FA = false"
              >
                Cancel
              </button>
            </div>
            <button
              v-else
              class="button button--flat"
              @click="showDisable2FA = true"
            >
              Disable 2FA
            </button>
          </div>

          <div v-else>
            <div v-if="!show2FASetup">
              <p style="margin-bottom: 1em; opacity: 0.85;">
                Protect your account by requiring an authentication code from your Authenticator app (Google Authenticator, Bitwarden, 1Password, etc.) on login.
              </p>
              <button class="button button--flat" @click="start2FASetup">
                Enable 2FA
              </button>
            </div>

            <div v-else style="display: flex; flex-direction: column; align-items: center; text-align: center;">
              <p style="margin-bottom: 0.75em;">Scan this QR code with your Authenticator app:</p>
              <div style="background: white; padding: 12px; border-radius: 8px; display: inline-block;">
                <qrcode-vue :value="totpURI" :size="180" level="H" />
              </div>
              <p style="margin-top: 0.75em; font-family: monospace; font-size: 0.9em; word-break: break-all;">
                Secret: <strong>{{ totpSecret }}</strong>
              </p>
              <input
                class="input input--block"
                type="text"
                maxlength="6"
                placeholder="Enter 6-digit code"
                v-model="totpVerificationCode"
                style="letter-spacing: 0.2em; text-align: center; font-size: 1.2em; margin-top: 1em; max-width: 240px;"
              />
              <div style="margin-top: 1em;">
                <button class="button button--flat" @click="confirm2FASetup">
                  Confirm & Enable
                </button>
                <button
                  class="button button--flat"
                  style="margin-left: 0.5em;"
                  @click="show2FASetup = false"
                >
                  Cancel
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useAuthStore } from "@/stores/auth";
import { useLayoutStore, type ThemeMode } from "@/stores/layout";
import { users as api } from "@/api";
import AceEditorTheme from "@/components/settings/AceEditorTheme.vue";
import Languages from "@/components/settings/Languages.vue";
import QrcodeVue from "qrcode.vue";
import { computed, inject, onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import { authMethod, noAuth } from "@/utils/constants";

const layoutStore = useLayoutStore();
const authStore = useAuthStore();
const { t } = useI18n();

const $showSuccess = inject<IToastSuccess>("$showSuccess")!;
const $showError = inject<IToastError>("$showError")!;

const selectedTheme = ref<ThemeMode>(layoutStore.theme);
const password = ref<string>("");
const passwordConf = ref<string>("");
const currentPassword = ref<string>("");
const isCurrentPasswordRequired = ref<boolean>(false);
const hideDotfiles = ref<boolean>(false);
const singleClick = ref<boolean>(false);
const redirectAfterCopyMove = ref<boolean>(false);
const dateFormat = ref<boolean>(false);
const locale = ref<string>("");
const aceEditorTheme = ref<string>("");

// 2FA state refs
const show2FASetup = ref<boolean>(false);
const showDisable2FA = ref<boolean>(false);
const totpSecret = ref<string>("");
const totpURI = ref<string>("");
const totpVerificationCode = ref<string>("");
const disable2FAPassword = ref<string>("");

const start2FASetup = async () => {
  try {
    const res = await api.generate2FA();
    totpSecret.value = res.secret;
    totpURI.value = res.uri;
    totpVerificationCode.value = "";
    show2FASetup.value = true;
  } catch (e: any) {
    $showError(e);
  }
};

const confirm2FASetup = async () => {
  if (!totpVerificationCode.value || totpVerificationCode.value.length !== 6) {
    $showError("Please enter the 6-digit verification code");
    return;
  }

  try {
    const res = await api.verify2FA(totpSecret.value, totpVerificationCode.value);
    if (res.success && authStore.user) {
      authStore.updateUser({ totpEnabled: true });
      show2FASetup.value = false;
      $showSuccess("Two-Factor Authentication successfully enabled!");
    }
  } catch (e: any) {
    $showError(e);
  }
};

const onDisable2FA = async () => {
  if (!disable2FAPassword.value) {
    $showError("Please enter your current password");
    return;
  }

  try {
    const res = await api.disable2FA(disable2FAPassword.value);
    if (res.success && authStore.user) {
      authStore.updateUser({ totpEnabled: false });
      showDisable2FA.value = false;
      disable2FAPassword.value = "";
      $showSuccess("Two-Factor Authentication disabled.");
    }
  } catch (e: any) {
    $showError(e);
  }
};

const passwordClass = computed(() => {
  const baseClass = "input input--block";

  if (password.value === "" && passwordConf.value === "") {
    return baseClass;
  }

  if (password.value === passwordConf.value) {
    return `${baseClass} input--green`;
  }

  return `${baseClass} input--red`;
});

onMounted(async () => {
  layoutStore.loading = true;
  if (authStore.user === null) return false;
  locale.value = authStore.user.locale;
  hideDotfiles.value = authStore.user.hideDotfiles;
  singleClick.value = authStore.user.singleClick;
  redirectAfterCopyMove.value = authStore.user.redirectAfterCopyMove;
  dateFormat.value = authStore.user.dateFormat;
  aceEditorTheme.value = authStore.user.aceEditorTheme;
  layoutStore.loading = false;
  isCurrentPasswordRequired.value = authMethod == "json";

  return true;
});

const updatePassword = async (event: Event) => {
  event.preventDefault();

  if (
    password.value !== passwordConf.value ||
    password.value === "" ||
    currentPassword.value === "" ||
    authStore.user === null
  ) {
    return;
  }

  try {
    const data = {
      ...authStore.user,
      id: authStore.user.id,
      password: password.value,
    };
    await api.update(data, ["password"], currentPassword.value);
    authStore.updateUser(data);
    $showSuccess(t("settings.passwordUpdated"));
  } catch (e: any) {
    $showError(e);
  } finally {
    password.value = passwordConf.value = "";
  }
};
const updateSettings = async (event: Event) => {
  event.preventDefault();

  try {
    if (authStore.user === null) throw new Error("User is not set!");

    const data = {
      ...authStore.user,
      id: authStore.user.id,
      locale: locale.value,
      hideDotfiles: hideDotfiles.value,
      singleClick: singleClick.value,
      redirectAfterCopyMove: redirectAfterCopyMove.value,
      dateFormat: dateFormat.value,
      aceEditorTheme: aceEditorTheme.value,
    };

    await api.update(data, [
      "locale",
      "hideDotfiles",
      "singleClick",
      "redirectAfterCopyMove",
      "dateFormat",
      "aceEditorTheme",
    ]);
    authStore.updateUser(data);
    $showSuccess(t("settings.settingsUpdated"));
  } catch (err) {
    if (err instanceof Error) {
      $showError(err);
    }
  }
};
</script>
