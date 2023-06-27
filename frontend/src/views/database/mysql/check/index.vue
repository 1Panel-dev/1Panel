<template>
    <el-dialog
        v-model="open"
        :title="$t('app.checkTitle')"
        width="50%"
        :close-on-click-modal="false"
        :destroy-on-close="true"
    >
        <el-row>
            <el-alert
                type="warning"
                :description="$t('app.deleteHelper', [$t('app.database')])"
                center
                show-icon
                :closable="false"
            />
            <el-col :span="12" :offset="6">
                <br />
                <el-descriptions border :column="1">
                    <el-descriptions-item>
                        <template #label>
                            <a href="javascript:void(0);" @click="toApp()">{{ $t('app.app') }}</a>
                        </template>
                        {{ installData.join(',') }}
                    </el-descriptions-item>
                </el-descriptions>
            </el-col>
        </el-row>
    </el-dialog>
</template>
<script lang="ts" setup>
import { ref } from 'vue';
import { useRouter } from 'vue-router';
const router = useRouter();

interface InstallRrops {
    items: Array<string>;
}
const installData = ref();
let open = ref(false);

const acceptParams = (props: InstallRrops) => {
    installData.value = props.items;
    open.value = true;
};

const toApp = () => {
    router.push({ name: 'AppInstalled' });
};

defineExpose({
    acceptParams,
});
</script>
