<template>
  <div id="login" :class="{ recaptcha: recaptcha }">
    <form @submit="submit">
      <img :src="logoURL" alt="File Browser" />
      <h1>{{ name }}</h1>
      <p v-if="reason != null" class="logout-message">
        {{ t(`login.logout_reasons.${reason}`) }}
      </p>
      <div v-if="error !== ''" class="wrong">{{ error }}</div>

      <template v-if="!requires2FA">
        <input
          autofocus
          class="input input--block"
          type="text"
          autocapitalize="off"
          v-model="username"
          :placeholder="t('login.username')"
        />
        <input
          class="input input--block"
          type="password"
          v-model="password"
          :placeholder="t('login.password')"
        />
        <input
          class="input input--block"
          v-if="createMode"
          type="password"
          v-model="passwordConfirm"
          :placeholder="t('login.passwordConfirm')"
        />
      </template>

      <template v-else>
        <p style="margin-bottom: 1em; font-size: 0.9em; opacity: 0.85;">
          🔒 Two-Factor Authentication (2FA)
        </p>
        <input
          autofocus
          class="input input--block"
          type="text"
          inputmode="numeric"
          pattern="[0-9]*"
          maxlength="6"
          v-model="totpCode"
          placeholder="6-digit 2FA Code"
          style="letter-spacing: 0.25em; text-align: center; font-size: 1.2em;"
        />
      </template>

      <div v-if="recaptcha" id="recaptcha"></div>
      <input
        class="button button--block"
        type="submit"
        :value="requires2FA ? 'Verify Code' : createMode ? t('login.signup') : t('login.submit')"
      />

      <p @click="requires2FA ? (requires2FA = false) : toggleMode()" v-if="requires2FA || signup">
        {{ requires2FA ? '← Back to Login' : createMode ? t("login.loginInstead") : t("login.createAnAccount") }}
      </p>
    </form>
  </div>
</template>

<script setup lang="ts">
import { StatusError } from "@/api/utils";
import * as auth from "@/utils/auth";
import {
  name,
  logoURL,
  recaptcha,
  recaptchaKey,
  signup,
} from "@/utils/constants";
import { inject, onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import { useRoute, useRouter } from "vue-router";

// Define refs
const createMode = ref<boolean>(false);
const requires2FA = ref<boolean>(false);
const totpCode = ref<string>("");
const error = ref<string>("");
const username = ref<string>("");
const password = ref<string>("");
const passwordConfirm = ref<string>("");

const route = useRoute();
const router = useRouter();
const { t } = useI18n({});
// Define functions
const toggleMode = () => {
  createMode.value = !createMode.value;
  requires2FA.value = false;
  totpCode.value = "";
};

const $showError = inject<IToastError>("$showError")!;

const reason = route.query["logout-reason"] ?? null;

const submit = async (event: Event) => {
  event.preventDefault();
  event.stopPropagation();
  error.value = "";

  const redirect = (route.query.redirect || "/files/") as string;

  let captcha = "";
  if (recaptcha) {
    captcha = window.grecaptcha.getResponse();

    if (captcha === "") {
      error.value = t("login.wrongCredentials");
      return;
    }
  }

  if (createMode.value) {
    if (password.value !== passwordConfirm.value) {
      error.value = t("login.passwordsDontMatch");
      return;
    }
  }

  try {
    if (createMode.value) {
      await auth.signup(username.value, password.value);
    }

    const result = await auth.login(
      username.value,
      password.value,
      captcha,
      totpCode.value
    );

    if (result && result.twoFactorRequired) {
      requires2FA.value = true;
      return;
    }

    router.push({ path: redirect });
  } catch (e: any) {
    // console.error(e);
    if (e instanceof StatusError) {
      if (e.status === 429) {
        error.value = "Too many failed attempts. Please wait a few minutes.";
      } else if (e.status === 409) {
        error.value = t("login.usernameTaken");
      } else if (e.status === 403) {
        error.value = t("login.wrongCredentials");
      } else if (e.status === 401) {
        error.value = "Invalid 2FA code. Please try again.";
      } else if (e.status === 400) {
        const match = e.message.match(/minimum length is (\d+)/);
        if (match) {
          error.value = t("login.passwordTooShort", { min: match[1] });
        } else {
          error.value = e.message;
        }
      } else {
        $showError(e);
      }
    }
  }
};

// Run hooks
onMounted(() => {
  if (!recaptcha) return;

  window.grecaptcha.ready(function () {
    window.grecaptcha.render("recaptcha", {
      sitekey: recaptchaKey,
    });
  });
});
</script>
