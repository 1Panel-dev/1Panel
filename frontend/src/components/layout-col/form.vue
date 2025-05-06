<template>
    <el-col :xs="24" :sm="span" :md="span" :lg="span" :xl="span">
        <slot />
    </el-col>
</template>

<script>
export default {
    props: {
        span: {
            type: Number,
            default: 10,
        },
    },
};
</script>
