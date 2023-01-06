<template>
    <div>
        <el-card>
            <el-divider direction="vertical" />
            <span style="font-size: 16px; font-weight: 700">{{ header }}</span>
            <div style="margin-top: 20px"><slot name="body" /></div>
        </el-card>
    </div>
</template>

<script setup lang="ts">
defineOptions({ name: 'CardWithHeader' });
defineProps({
    header: String,
});
</script>

<style scoped lang="scss">
.el-divider--vertical {
    display: inline-block;
    width: 1px;
    height: 2.5em;
    margin: 0px 8px;
    vertical-align: middle;
    position: relative;
    border-left: 5px #0054e1 var(--el-border-style);
}
</style>
