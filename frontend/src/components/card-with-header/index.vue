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
            &::before {
                position: absolute;
                top: 4px;
                left: -13px;
                width: 4px;
                height: 14px;
                content: '';
                background: $primary-color;
                border-radius: 10px;
            }
        }
        // span:first-child {
        //     background: $primary-color;
        //     border-radius: 7px;
        //     flex: none;
        //     order: 0;
        //     flex-grow: 0;
        //     padding: 3px;
        // }

        // span:nth-child(2) {
        //     font-size: 18px;
        //     font-weight: 700;
        //     margin-left: 10px;
        // }
    }

    div:nth-child(2) {
        margin-top: 20px;
    }
}
</style>
