<template>
    <div>
        <div class="search-button">
            <el-input
                clearable
                v-model="searchInfo"
                @clear="search()"
                suffix-icon="Search"
                @change="search()"
                :disabled="props.disabled"
                :placeholder="props.placeholder || $t('commons.button.search')"
            ></el-input>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue';
import { useRoute } from 'vue-router';
defineOptions({ name: 'TableSearch' });

const emit = defineEmits(['search', 'update:searchName']);
const searchInfo = ref();
const route = useRoute();
const props = defineProps({
    placeholder: String,
    disabled: {
        type: Boolean,
        default: false,
    },
    searchName: {
        type: [String, Number],
        default: undefined,
    },
});

watch(
    () => props.searchName,
    (newVal) => {
        if (searchInfo.value !== newVal) {
            searchInfo.value = newVal;
        }
    },
    { immediate: true },
);

// Also discard text typed without pressing Enter when leaving a cached tab.
watch(
    () => route.path,
    () => {
        searchInfo.value = '';
        emit('update:searchName', '');
    },
    { flush: 'sync' },
);

const search = () => {
    emit('update:searchName', searchInfo.value);
    emit('search');
};
</script>
