<template>
    <el-page-header :content="header" @back="jump">
        <template v-if="slots.buttons" #content>
            <span>{{ header }}</span>
            <span v-if="!mobile">
                <el-divider direction="vertical" />
                <slot name="buttons"></slot>
            </span>
        </template>
    </el-page-header>
    <template v-if="slots.buttons && mobile">
        <slot name="buttons"></slot>
    </template>
</template>

<script setup lang="ts">
import { computed, inject, useSlots } from 'vue';
import { useRouter } from 'vue-router';
import { GlobalStore } from '@/store';

const globalStore = GlobalStore();
const slots = useSlots();
const router = useRouter();
const props = defineProps({
    path: String,
    name: String,
    to: Object,
    header: String,
    reload: Boolean,
});
function jump() {
    const { path, name, to, reload } = props;
    if (reload) {
        reloadPage();
    }
    if (path) {
        router.push(path);
    }
    if (name) {
        router.push({ name: name });
    }
    if (to) {
        router.push(to);
    }
}

let reloadPage: Function = inject('reload');

const mobile = computed(() => {
    return globalStore.isMobile();
});
</script>
