<template>
  <div class="card contain-card mb-3">
    <div class="card-header">
      {{ group.id ? `${$t('update')} ${group.name}` : $t('group_create') }}
      <transition name="slide-fade">
        <button @click="removeEdit" v-if="group.id" class="btn float-right btn-danger btn-sm">
          {{ $t('close') }}
        </button>
      </transition>
    </div>
    <div class="card-body">
      <form @submit.prevent="saveGroup">
        <div class="form-group row">
          <label for="title" class="col-sm-4 col-form-label">{{ $t('group') }} {{ $t('name') }}</label>
          <div class="col-sm-8">
            <input v-model="group.name" type="text" class="form-control" id="title" placeholder="Group Name" required />
          </div>
        </div>
        <div class="form-group row">
          <label for="switch-group-public" class="col-sm-4 col-form-label text-capitalize"
            >{{ $t('public') }} {{ $t('group') }}</label
          >
          <div class="col-md-8 col-xs-12 mt-1">
            <span @click="group.public = !!group.public" class="switch float-left">
              <input v-model="group.public" type="checkbox" class="switch" id="switch-group-public" :checked="group.public" />
              <label for="switch-group-public">{{ $t('group_public_desc') }}</label>
            </span>
          </div>
        </div>
        <div class="form-group row">
          <div class="col-6">
            <button
              @click.prevent="cancelEdit"
              type="button"
              class="btn btn-block btn-outline-secondary"
            >
              Cancel
            </button>
          </div>
          <div class="col-6">
            <button
              @click.prevent="saveGroup"
              type="submit"
              :disabled="loading || group.name === ''"
              class="btn btn-block"
              :class="{ 'btn-primary': !group.id, 'btn-secondary': group.id }"
            >
              {{ loading ? 'Loading...' : group.id ? $t('group_update') : $t('group_create') }}
            </button>
          </div>
        </div>
        <div class="alert alert-danger d-none" id="alerter" role="alert"></div>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from "vue";
import Api from "@/API";
import { useMainStore } from "@/stores/main";

const props = defineProps({
	in_group: {
		type: Object,
		default: () => ({}),
	},
	edit: {
		type: Function,
		default: () => {},
	},
});

const emit = defineEmits(["cancel"]);

const store = useMainStore();
const loading = ref(false);
const group = ref({
	name: "",
	public: true,
});

watch(
	() => props.in_group,
	(val) => {
		if (val && Object.keys(val).length > 0) {
			group.value = { ...val };
		}
	},
	{ immediate: true },
);

function removeEdit() {
	group.value = { name: "", public: true };
	props.edit(false);
}

function cancelEdit() {
	group.value = { name: "", public: true };
	props.edit(false);
	emit("cancel");
}

async function saveGroup() {
	loading.value = true;
	if (group.value.id) {
		await updateGroup();
	} else {
		await createGroup();
	}
	loading.value = false;
}

async function createGroup() {
	const g = group.value;
	const data = { name: g.name, public: g.public };
	await Api.group_create(data);
	await update();
	group.value = { name: "", public: true };
	emit("cancel");
}

async function updateGroup() {
	const g = group.value;
	const data = { id: g.id, name: g.name, public: g.public };
	await Api.group_update(data);
	await update();
	props.edit(false);
}

async function update() {
	const groups = await Api.groups();
	store.setGroups(groups);
}
</script>

<style scoped></style>
