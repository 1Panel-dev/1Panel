<template>
    <el-card class="config-card">
        <span class="web-tag" v-if="website">
            <el-tooltip :content="$t('xpack.waf.websiteHelper')" placement="bottom">
                <el-tag type="primary" size="small">{{ $t('menu.website') }}</el-tag>
            </el-tooltip>
        </span>
        <div class="config-header">
            <span>
                {{ header }}
            </span>
            <slot name="header-r" />
        </div>

        <el-text type="info">{{ description }}</el-text>
        <div class="config-content">
            <slot name="content-r" />
        </div>
        <div class="config-form">
            <slot name="content-form" />
        </div>
    </el-card>
</template>

<script setup lang="ts">
defineOptions({ name: 'ConfigCard' });
defineProps({
    header: String,
    description: String,
    website: Boolean,
});
</script>

<style lang="scss" scoped>
.config-card {
    cursor: pointer;

    .web-tag {
        position: absolute;
        left: 0px;
        top: 0px;
    }

    .config-header {
        margin-bottom: 18px;
        display: flex;
        justify-content: space-between;
        align-items: center;
        span {
            font-weight: normal;
            font-size: 18px;
        }
    }

    .config-content {
        display: flex;
        justify-content: flex-end;
    }

    .config-form {
        margin-top: 20px;
    }
}
.config-card {
    border: var(--panel-border) !important;
    &:hover {
        cursor: pointer;
        border: 1px solid var(--el-color-primary) !important;
    }
}
</style>
