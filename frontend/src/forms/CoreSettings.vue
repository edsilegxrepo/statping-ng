<template>
  <form @submit.prevent="saveSettings">
    <div class="card">
      <div class="card-header">Statping-ng Settings</div>
      <div class="card-body">
        <div class="form-group">
          <label>{{ $t('project_name') }}</label>
          <input v-model="coreData.name" type="text" class="form-control" placeholder="Great Uptime" id="project" />
        </div>

        <div class="form-group">
          <label>{{ $t('description') }}</label>
          <input
            v-model="coreData.description"
            type="text"
            class="form-control"
            placeholder="Great Uptime"
            id="description"
          />
        </div>

        <div class="form-group">
          <label>{{ $t('domain') }}</label>
          <input v-model="coreData.domain" type="url" class="form-control" id="domain" />
        </div>

        <div class="form-group">
          <label>{{ $t('footer') }}</label>
          <textarea v-model="coreData.footer" rows="4" class="form-control" id="footer"></textarea>
          <small class="form-text text-muted">{{ $t('footer_notes') }}</small>
        </div>

        <div class="form-group">
          <label>{{ $t('language') }}</label>
          <select v-model="coreData.language" class="form-control">
            <option value="en">English</option>
            <option value="es">Spanish</option>
            <option value="fr">French</option>
            <option value="ru">Russian</option>
            <option value="de">German</option>
            <option value="cs">Czech</option>
            <option value="ja">Japanese</option>
            <option value="it">Italian</option>
            <option value="ko">Korean</option>
            <option value="zh">Chinese</option>
            <option value="sv">Swedish</option>
          </select>
        </div>

        <div class="form-group mt-3">
          <label>Session Timeout (minutes)</label>
          <input
            v-model.number="coreData.session_timeout"
            type="number"
            class="form-control"
            min="5"
            max="10080"
            placeholder="720"
            id="session_timeout"
          />
          <small class="form-text text-muted">
            How long until users are logged out (default: 720 = 12 hours, max: 10080 = 7 days)
          </small>
        </div>
      </div>
      <div class="card-footer">
        <button
          @click.prevent="saveSettings"
          id="save_core"
          type="submit"
          class="btn btn-primary btn-block"
          :disabled="loading"
        >
          <font-awesome-icon v-if="loading" icon="circle-notch" class="mr-2" spin />{{ $t('save_settings') }}
        </button>
      </div>
    </div>
  </form>
</template>

<script setup>
import { computed, ref } from "vue";
import { useI18n } from "vue-i18n";
import Api from "@/API";
import { useMainStore } from "@/stores/main";

const store = useMainStore();
const { locale } = useI18n();

const loading = ref(false);

const coreData = computed(() => store.core);

async function saveSettings() {
	loading.value = true;
	const c = coreData.value;
	await Api.core_save(c);
	store.setCore(c);
	locale.value = c.language || "en";
	loading.value = false;
}
</script>

<style scoped></style>
