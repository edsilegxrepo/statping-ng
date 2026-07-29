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

        <div class="form-group row">
          <div class="col-8 col-sm-9">
            <label>{{ $t('domain') }}</label>
            <input v-model="coreData.domain" type="url" class="form-control" id="domain" />
          </div>
          <div class="col-4 col-sm-3 mt-sm-1 mt-0">
            <label class="d-inline d-sm-none">{{ $t('enable_cdn') }}</label>
            <label class="d-none d-sm-block">{{ $t('enable_cdn') }}</label>
            <span @click="coreData.using_cdn = !!coreData.using_cdn" class="switch" id="using_cdn">
              <input
                v-model="coreData.using_cdn"
                type="checkbox"
                name="using_cdn"
                class="switch"
                id="switch-normal"
                :checked="coreData.using_cdn"
              />
              <label for="switch-normal"></label>
            </span>
          </div>
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

        <div class="form-group row mt-3">
          <label class="col-sm-10 col-form-label">{{ $t('send_reports') }}</label>
          <div class="col-sm-2 float-right">
            <span @click="coreData.allow_reports = !!coreData.allow_reports" class="switch" id="allow_report">
              <input
                v-model="coreData.allow_reports"
                type="checkbox"
                name="allow_report"
                class="switch"
                id="switch_allow_report"
                :checked="coreData.allow_reports"
              />
              <label for="switch_allow_report"></label>
            </span>
          </div>
          <div class="col-12">
            <small>{{ $t('send_reports_desc') }}</small>
          </div>
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
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useMainStore } from '@/stores/main'
import Api from '@/API'

const store = useMainStore()
const { locale } = useI18n()

const loading = ref(false)

const coreData = computed(() => store.core)

async function saveSettings() {
  loading.value = true
  const c = coreData.value
  await Api.core_save(c)
  store.setCore(c)
  locale.value = c.language || 'en'
  loading.value = false
}
</script>

<style scoped></style>
