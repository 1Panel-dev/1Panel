<template>
    <el-dialog
        v-model="open"
        :title="$t('app.checkTitle')"
        width="50%"
        :close-on-click-modal="false"
        :destroy-on-close="true"
    >
        <el-row>
            <el-alert type="warning" :description="$t('app.deleteHelper')" center show-icon :closable="false" />
            <el-col :span="12" :offset="6">
                <br />
                <el-descriptions border :column="1">
                    <el-descriptions-item v-for="(item, key) in map" :key="key" :label="$t('app.' + item[0])">
                        {{ map.get(item[0]).toString() }}
                    </el-descriptions-item>
                </el-descriptions>
            </el-col>
        </el-row>
    </el-dialog>
</template>
<script lang="ts" setup>
import { App } from '@/api/interface/app';
import { ref } from 'vue';

interface InstallRrops {
    items: App.AppInstallResource[];
}
const installData = ref<InstallRrops>({
    items: [],
});
let open = ref(false);
let map = new Map();

const acceptParams = (props: InstallRrops) => {
    map.clear();
    installData.value.items = [];
    installData.value.items = props.items;
    installData.value.items.forEach((item) => {
        if (map.has(item.type)) {
            const array = map.get(item.type);
            array.push(item.name);
            map.set(item.type, array);
        } else {
            map.set(item.type, [item.name]);
        }
    });
    open.value = true;
};

defineExpose({
    acceptParams,
});
</script>
