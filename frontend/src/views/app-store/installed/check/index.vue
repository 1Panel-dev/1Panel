<template>
    <el-dialog v-model="open" :title="$t('app.checkTitle')" width="50%" :close-on-click-modal="false">
        <el-row>
            <el-alert
                type="warning"
                :description="$t('app.deleteHelper', [$t('app.app')])"
                center
                show-icon
                :closable="false"
            />
            <el-col :span="20" :offset="2" v-if="open">
                <br />
                <el-descriptions border :column="1">
                    <el-descriptions-item v-for="(item, key) in map" :key="key">
                        <template #label>
                            <a href="javascript:void(0);" @click="toPage(item[0])">{{ $t('app.' + item[0]) }}</a>
                        </template>
                        <span style="word-break: break-all">
                            {{ map.get(item[0]).toString() }}
                        </span>
                    </el-descriptions-item>
                </el-descriptions>
            </el-col>
        </el-row>
    </el-dialog>
</template>
<script lang="ts" setup>
import { App } from '@/api/interface/app';
import { ref } from 'vue';
import { useRouter } from 'vue-router';
const router = useRouter();

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

const toPage = (key: string) => {
    if (key === 'app') {
        open.value = false;
    }
    if (key === 'website') {
        router.push({ name: 'Website' });
    }
    if (key === 'database') {
        router.push({ name: 'MySQL' });
    }
};

defineExpose({
    acceptParams,
});
</script>
