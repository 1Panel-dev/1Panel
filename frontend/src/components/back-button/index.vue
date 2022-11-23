<template>
    <el-page-header :content="header" @back="jump" />
</template>

<script setup lang="ts">
import { inject } from 'vue';
import { useRouter } from 'vue-router';
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
</script>
