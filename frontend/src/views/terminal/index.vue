<template>
    <LayoutContent :header="$t('menu.terminal')">
        <div>
            <el-tabs v-model="terminalValue">
                <el-tab-pane :key="item.name" v-for="item in terminalTabs" :label="item.title" :name="item.name">
                    <iframe id="iframeTerminal" name="iframeTerminal" width="100%" frameborder="0" :src="item.src" />
                </el-tab-pane>
            </el-tabs>
        </div>
    </LayoutContent>
</template>

<script setup lang="ts">
import { onMounted, ref, nextTick } from 'vue';
import LayoutContent from '@/layout/layout-content.vue';

const terminalValue = ref();
const terminalTabs = ref([]) as any;

function changeFrameHeight() {
    let ifm = document.getElementById('iframeTerminal') as HTMLInputElement | null;
    if (ifm) {
        ifm.style.height = document.documentElement.clientHeight - 280 + 'px';
    }
}

window.onresize = function () {
    changeFrameHeight();
};

onMounted(() => {
    terminalTabs.value.push({ name: '本地服务器', title: '本地服务器', src: 'http://localhost:8080' });
    terminalValue.value = '本地服务器';
    nextTick(() => {
        changeFrameHeight();
    });
});
</script>
<style lang="scss" scoped></style>
