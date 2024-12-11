<template>
    <el-dialog
        v-model="open"
        :title="$t('app.checkTitle')"
        width="50%"
        :close-on-click-modal="false"
        :destroy-on-close="true"
    >
        <el-row>
            <el-alert type="warning" :description="$t('website.deleteHelper')" center show-icon :closable="false" />
            <el-col :span="24">
                <br />
                <el-table :data="items" style="width: 100%">
                    <el-table-column prop="name" :label="$t('app.installName')" />
                    <el-table-column prop="appName" :label="$t('app.appName')" />
                    <el-table-column prop="version" :label="$t('app.version')" />
                    <el-table-column prop="status" :label="$t('app.status')" />
                </el-table>
            </el-col>
        </el-row>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="toPage">
                    {{ $t('website.toApp') }}
                </el-button>
            </span>
        </template>
    </el-dialog>
</template>
<script lang="ts" setup>
import { Website } from '@/api/interface/website';
import { ref } from 'vue';
import { useRouter } from 'vue-router';
const router = useRouter();

interface InstallProps {
    items: Website.CheckRes[];
}

let open = ref(false);
let items = ref([]);

const acceptParams = async (props: InstallProps) => {
    items.value = props.items;
    open.value = true;
};

const handleClose = () => {
    open.value = false;
};

const toPage = () => {
    router.push({ name: 'AppInstalled' });
};

defineExpose({
    acceptParams,
});
</script>
