<template>
    <div>
        <LayoutContent title="vLLM">
            <template #leftToolBar>
                <el-button type="primary" @click="goLicense">
                    {{ $t('license.levelUpPro') }}
                </el-button>
            </template>
            <template #prompt>
                <el-alert type="info" :closable="false">
                    <template #title>
                        <span>{{ $t('license.vllm') }}</span>
                    </template>
                </el-alert>
            </template>
            <template #main>
                <div class="vllm-placeholder">
                    <div class="vllm-placeholder__intro">
                        <div class="vllm-placeholder__title">{{ $t('license.introduce') }}</div>
                        <div class="vllm-placeholder__desc">{{ $t('license.vllm') }}</div>
                    </div>
                    <el-row :gutter="16">
                        <el-col v-for="item in features" :key="item.title" :xs="24" :sm="12" :xl="6">
                            <el-card class="feature-card" shadow="hover">
                                <template #header>
                                    <div class="feature-card__header">
                                        <svg-icon class="feature-card__icon" :iconName="item.icon" />
                                        <span>{{ $t(item.title) }}</span>
                                    </div>
                                </template>
                                <div class="feature-card__content">{{ $t(item.content) }}</div>
                            </el-card>
                        </el-col>
                    </el-row>
                </div>
            </template>
        </LayoutContent>
    </div>
</template>

<script setup lang="ts">
import { routerToName } from '@/utils/router';

const features = [
    { icon: 'p-ai-1', title: 'xpack.helper.vllmTitle1', content: 'xpack.helper.vllmContent1' },
    { icon: 'p-ai-2', title: 'xpack.helper.vllmTitle2', content: 'xpack.helper.vllmContent2' },
    { icon: 'p-ai-3', title: 'xpack.helper.vllmTitle3', content: 'xpack.helper.vllmContent3' },
    { icon: 'p-ai-4', title: 'xpack.helper.vllmTitle4', content: 'xpack.helper.vllmContent4' },
];

const goLicense = () => {
    routerToName('License');
};
</script>

<style scoped lang="scss">
.vllm-placeholder {
    padding-bottom: 16px;
}

.vllm-placeholder__intro {
    max-width: 760px;
    margin-bottom: 20px;
}

.vllm-placeholder__title {
    font-size: 28px;
    font-weight: 500;
    line-height: 1.4;
}

.vllm-placeholder__desc {
    margin-top: 12px;
    color: var(--el-text-color-regular);
    line-height: 1.7;
}

.feature-card {
    height: 100%;
    border: var(--panel-border);
}

.feature-card__header {
    display: flex;
    align-items: center;
    gap: 10px;
    font-size: 16px;
    font-weight: 500;
}

.feature-card__icon {
    font-size: 18px;
}

.feature-card__content {
    color: var(--el-text-color-regular);
    line-height: 1.7;
}
</style>
