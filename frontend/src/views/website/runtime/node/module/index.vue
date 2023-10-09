<template>
    <el-drawer :close-on-click-modal="false" v-model="open" size="50%">
        <template #header>
            <DrawerHeader :header="$t('runtime.moduleManager')" :back="handleClose" />
        </template>
        <el-row>
            <el-col :span="24">
                <ComplexTable :data="data" @search="search()" :height="650">
                    <el-table-column :label="$t('commons.table.name')" prop="name" min-width="100px"></el-table-column>
                    <el-table-column :label="$t('container.version')" prop="version" width="80px"></el-table-column>
                    <el-table-column
                        :label="$t('commons.table.protocol')"
                        prop="license"
                        width="120px"
                    ></el-table-column>
                    <el-table-column
                        :label="$t('container.description')"
                        fix
                        min-width="120px"
                        prop="description"
                    ></el-table-column>
                </ComplexTable>
            </el-col>
        </el-row>
    </el-drawer>
</template>
<script setup lang="ts">
import { ref } from 'vue';
import { GetNodeModules } from '@/api/modules/runtime';

interface NoodeRrops {
    packageManager: string;
    id: number;
}

const open = ref(false);
const id = ref(0);
const data = ref([]);

const acceptParams = async (props: NoodeRrops) => {
    id.value = props.id;
    data.value = [];
    search();
    open.value = true;
};

const search = async () => {
    try {
        const res = await GetNodeModules({ ID: id.value });
        data.value = res.data;
        console.log(res);
    } catch (error) {}
};

const handleClose = () => {
    open.value = false;
};

defineExpose({
    acceptParams,
});
</script>

<style scoped lang="scss">
.table-border {
    border: 1px solid var(--el-border-color);
}
</style>
