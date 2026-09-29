<template>
    <div v-if="sections.length" class="device-details">
        <section v-for="section in sections" :key="section.id" class="detail-section">
            <h3>{{ section.title }}</h3>
            <div class="detail-grid">
                <div v-for="item in section.items" :key="item.label" class="detail-item">
                    <div class="detail-label">
                        <span>{{ item.label }}</span>
                        <el-tooltip v-if="item.help" placement="top">
                            <template #content>
                                <span class="detail-help">{{ item.help }}</span>
                            </template>
                            <el-icon tabindex="0" :aria-label="item.help"><InfoFilled /></el-icon>
                        </el-tooltip>
                    </div>
                    <el-tooltip v-if="item.ellipsis" :content="item.value" placement="top">
                        <strong class="detail-ellipsis" tabindex="0">{{ item.value }}</strong>
                    </el-tooltip>
                    <strong v-else>{{ item.value }}</strong>
                </div>
            </div>
        </section>
    </div>
</template>

<script lang="ts" setup>
import { InfoFilled } from '@element-plus/icons-vue';

export interface DetailSection {
    id: string;
    title: string;
    items: { label: string; value?: string; help?: string; ellipsis?: boolean }[];
}

defineProps<{ sections: DetailSection[] }>();
</script>

<style lang="scss" scoped>
.detail-help {
    white-space: pre-line;
}
.device-details {
    display: grid;
    gap: 24px;
}
.detail-section {
    & + & {
        padding-top: 24px;
        border-top: 1px solid var(--el-border-color-lighter);
    }
    h3 {
        margin: 0 0 16px;
        color: var(--el-text-color-primary);
        font-size: 14px;
        font-weight: 600;
    }
}
.detail-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(min(180px, 100%), 1fr));
    gap: 16px 20px;
}
.detail-item {
    min-width: 0;
    .detail-label {
        display: flex;
        align-items: baseline;
        gap: 6px;
        margin-bottom: 5px;
        color: var(--el-text-color-secondary);
        font-size: 12px;
        line-height: 18px;
        overflow-wrap: anywhere;
    }
    strong {
        display: block;
        color: var(--el-text-color-primary);
        font-size: 13px;
        font-weight: 500;
        font-variant-numeric: tabular-nums;
        line-height: 20px;
        overflow-wrap: anywhere;
    }
    .detail-ellipsis {
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }
}
</style>
