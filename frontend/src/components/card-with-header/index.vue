<template>
    <div>
        <el-card :style="{ height: height }" class="home-card">
            <div class="header">
                <div class="header-left">
                    <span class="header-span">{{ header }}</span>
                    <slot name="header-l" />
                </div>
                <div class="header-right">
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
});
</script>

<style scoped lang="scss">
.home-card {
    .header {
        display: flex;
        justify-content: space-between;
        align-items: center;

        .header-left {
            display: flex;
            align-items: center;
            gap: 12px;

            .header-span {
                position: relative;
                font-size: 16px;
                font-weight: 500;
                margin-left: 18px;
                display: flex;
                align-items: center;

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
        }
    }

    .body-content {
        margin-top: 20px;
    }
}
</style>
