<template>
    <!-- <el-icon class="back-button" @click="jump">
        <Back />
    </el-icon> -->
    <el-page-header :content="header" @back="jump" />
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router';
const router = useRouter();
const props = defineProps({
    path: String,
    name: String,
    to: Object,
    header: String,
});
function jump() {
    const { path, name, to } = props;
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
</script>

<style lang="scss">
// .back-button {
//     cursor: pointer;
//     margin-right: 10px;
//     font-weight: 600;

//     &:active {
//         transform: scale(0.85);
//     }
// }
</style>
