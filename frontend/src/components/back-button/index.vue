<template>
    <el-page-header :content="header" @back="jump">
        <template v-if="slots.buttons" #content>
            <slot name="buttons"></slot>
        </template>
    </el-page-header>
</template>

<script setup lang="ts">
import { routerToName, routerToPath } from '@/utils/router';
import { inject, useSlots } from 'vue';
import { useRouter } from 'vue-router';

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
        routerToPath(path);
    }
    if (name) {
        routerToName(name);
    }
    if (to) {
        router.push(to);
    }
}

let reloadPage: Function = inject('reload');
</script>
