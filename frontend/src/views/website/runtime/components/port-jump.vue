<template>
    <div v-if="row.port != ''">
        <span v-for="(port, index) in row.exposedPorts" :key="index">
            <el-button icon="Position" plain size="small" @click="jump(port.hostPort, 'http')">
                {{ port.hostIP }}:{{ port.hostPort }}->{{ port.containerPort }}
            </el-button>
        </span>
    </div>
</template>

<script setup>
defineProps({
    row: {
        type: Object,
        required: true,
    },
    jump: {
        type: Function,
        required: true,
    },
});
</script>
