<template>
    <div>
        <el-card :style="{ height: height }" class="home-card">
            <div class="header">
                <span></span>
                <span>{{ header }}</span>
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
        span:first-child {
            background: $primary-color;
            border-radius: 7px;
            flex: none;
            order: 0;
            flex-grow: 0;
            padding: 3px;
        }

        span:nth-child(2) {
            font-size: 18px;
            font-weight: 700;
            margin-left: 10px;
        }
    }

    div:nth-child(2) {
        margin-top: 20px;
    }
}
</style>
