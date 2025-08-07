<template>
    <div>
        <el-row>
            <el-col :span="22" :offset="1">
                <el-descriptions border :column="1">
                    <el-descriptions-item v-for="(item, key) in data" :label="item.key" :key="key">
                        {{ item.value }}
                    </el-descriptions-item>
                </el-descriptions>
            </el-col>
        </el-row>
    </div>
</template>

<script lang="ts" setup>
import { getFPMStatus } from '@/api/modules/runtime';

const props = defineProps({
    id: {
        type: Number,
        default: 0,
    },
});
const data = ref([]);

const getStatus = async () => {
    try {
        const res = await getFPMStatus(props.id);
        data.value = res.data || [];
    } catch (error) {}
};

onMounted(() => {
    getStatus();
});
</script>
