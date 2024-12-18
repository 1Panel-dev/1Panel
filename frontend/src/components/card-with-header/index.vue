<template>
    <div>
        <el-card :style="{ height: height }" class="home-card">
            <div class="header">
                <span>{{ header }}</span>
                <slot name="header-r" />
            </div>

            <div>
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
        span {
            position: relative;
            font-size: 16px;
            font-weight: 500;
            margin-left: 18px;
            display: flex;
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

    div:nth-child(2) {
        margin-top: 20px;
    }
}
</style>
