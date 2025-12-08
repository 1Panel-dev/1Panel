<template>
    <el-drawer v-model="open" :title="title" size="50%" :before-close="handleClose" destroy-on-close>
        <div v-loading="loading">
            <div class="mb-4 flex items-center justify-between">
                <el-tag v-if="isTruncated" type="warning">
                    {{ $t('file.previewTruncated') }}
                </el-tag>
                <el-button @click="onDownload" icon="Download" :disabled="!hasContent">
                    {{ $t('commons.button.download') }}
                </el-button>
            </div>
            <div class="log-container" ref="logContainer" @scroll="onScroll">
                <div class="log-spacer" :style="{ height: `${totalHeight}px` }"></div>
                <div class="log-viewport" :style="{ transform: `translateY(${offsetY}px)` }">
                    <div
                        v-for="(line, index) in visibleLines"
                        :key="`${startIndex + index}-${line}`"
                        class="log-item"
                        :style="{ height: `${lineHeight}px` }"
                    >
                        <span class="line-content">{{ line }}</span>
                    </div>
                </div>
                <div v-if="lines.length === 0 && !loading" class="empty-content">
                    {{ $t('file.previewEmpty') }}
                </div>
            </div>
        </div>
    </el-drawer>
</template>

<script lang="ts" setup>
import { ref, computed, nextTick } from 'vue';
import { getPreviewContent } from '@/api/modules/files';
import { downloadFile } from '@/utils/util';
import { GlobalStore } from '@/store';
import i18n from '@/lang';

const globalStore = GlobalStore();

interface PreviewProps {
    path: string;
    name: string;
}

const open = ref(false);
const loading = ref(false);
const title = ref('');
const filePath = ref('');
const lines = ref<string[]>([]);
const isTruncated = ref(false);
const hasContent = ref(false);

const logContainer = ref<HTMLElement | null>(null);
const lineHeight = 23;
const scrollTop = ref(0);

const totalHeight = computed(() => lines.value.length * lineHeight);

const visibleCount = computed(() => {
    const buffer = 5;
    return Math.ceil(600 / lineHeight) + buffer * 2;
});

const startIndex = computed(() => {
    const buffer = 5;
    const index = Math.floor(scrollTop.value / lineHeight) - buffer;
    return Math.max(0, index);
});

const endIndex = computed(() => {
    return Math.min(lines.value.length, startIndex.value + visibleCount.value);
});

const visibleLines = computed(() => {
    return lines.value.slice(startIndex.value, endIndex.value);
});

const offsetY = computed(() => {
    return startIndex.value * lineHeight;
});

const onScroll = () => {
    if (!logContainer.value) return;
    scrollTop.value = logContainer.value.scrollTop;
};

const handleClose = () => {
    open.value = false;
    lines.value = [];
};

const onDownload = () => {
    downloadFile(filePath.value, globalStore.currentNode);
};

const acceptParams = async (props: PreviewProps) => {
    title.value = i18n.global.t('commons.button.preview') + ' - ' + props.name;
    filePath.value = props.path;
    lines.value = [];
    isTruncated.value = false;
    hasContent.value = false;
    open.value = true;
    loading.value = true;

    try {
        const res = await getPreviewContent({ path: props.path });
        if (res.data.content) {
            lines.value = res.data.content.split('\n');
            hasContent.value = true;
            if (res.data.size > 10 * 1024 * 1024) {
                isTruncated.value = true;
            }
        }
        nextTick(() => {
            if (logContainer.value) {
                logContainer.value.scrollTop = logContainer.value.scrollHeight;
            }
        });
    } catch (error) {
        console.error('Preview error:', error);
    } finally {
        loading.value = false;
    }
};

defineExpose({ acceptParams });
</script>

<style lang="scss" scoped>
.log-container {
    height: calc(100vh - 200px);
    overflow-y: auto;
    overflow-x: auto;
    position: relative;
    background-color: var(--panel-logs-bg-color);
    border-radius: 4px;
}

.log-spacer {
    position: relative;
    width: 100%;
}

.log-viewport {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    will-change: transform;
    width: max-content;
    min-width: 100%;
}

.log-item {
    min-width: 100%;
    padding: 2px 8px;
    color: #f5f5f5;
    box-sizing: border-box;
    white-space: pre;
}

.line-content {
    font-family: 'Courier New', Courier, monospace;
    font-size: 13px;
}

.empty-content {
    padding: 20px;
    text-align: center;
    color: #9ca3af;
}
</style>
