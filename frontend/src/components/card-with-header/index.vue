<template>
    <div :class="{ 'fill-height': fill }">
        <el-card :style="{ height: height }" class="home-card">
            <div class="header">
                <div class="header-left flex flex-wrap gap-3">
                    <span class="header-span">{{ header }}</span>
                    <slot name="header-l" />
                </div>
                <div class="header-right flex flex-wrap gap-3">
                    <slot name="header-r" />
                </div>
            </div>
            <div class="body-content">
                <slot name="body" />
            </div>
        </el-card>
    </div>
</template>

<script setup lang="ts">
defineOptions({ name: 'CardWithHeader' });
defineProps({
    header: String,
    height: String,
    fill: Boolean,
});
</script>

<style scoped lang="scss">
.fill-height {
    display: flex;
    flex: 1;
    min-height: 0;

    .home-card {
        flex: 1;
        min-height: 0;

        :deep(.el-card__body) {
            box-sizing: border-box;
            display: flex;
            flex-direction: column;
            height: 100%;
        }

        .header {
            flex-shrink: 0;
        }

        .body-content {
            flex: 1;
            min-height: 0;
        }
    }
}

.home-card {
    .header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        gap: 12px;
        min-width: 0;

        .header-left {
            display: flex;
            align-items: center;
            gap: 12px;
            min-width: 0;

            .header-span {
                position: relative;
                font-size: 16px;
                font-weight: 500;
                margin-left: 18px;
                display: flex;
                align-items: center;
                min-width: 0;
                overflow-wrap: anywhere;

                &::before {
                    position: absolute;
                    top: 50%;
                    transform: translateY(-50%);
                    left: -13px;
                    width: 4px;
                    height: 14px;
                    content: '';
                    background: $primary-color;
                    border-radius: 10px;
                }
            }
        }

        .header-right {
            display: flex;
            align-items: center;
            min-width: 0;
        }
    }

    .body-content {
        margin-top: 20px;
    }
}

@media (max-width: 767px) {
    .home-card .header {
        flex-wrap: wrap;
        align-items: flex-start;

        .header-left {
            flex: 1 1 200px;
            max-width: 100%;
        }

        .header-right {
            flex: 0 1 auto;
            max-width: 100%;
            margin-left: auto;
        }

        .header-right:empty {
            display: none;
        }
    }
}
</style>
