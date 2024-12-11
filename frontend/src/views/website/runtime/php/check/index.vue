<template>
    <el-dialog v-model="open" :title="$t('app.checkTitle')" width="40%" :close-on-click-modal="false">
        <el-row>
            <el-col :span="20" :offset="2" v-if="open">
                <el-alert
                    type="error"
                    :description="$t('app.deleteHelper', [$t('runtime.runtime')])"
                    center
                    show-icon
                    :closable="false"
                />
                <el-descriptions border :column="1" class="mt-5">
                    <el-descriptions-item
                        v-for="(item, key) in map"
                        :key="key"
                        label-class-name="check-label"
                        class-name="check-content"
                        min-width="60px"
                    >
                        <template #label>
                            <a href="javascript:void(0);" class="check-label-a" @click="toPage(item[0])">
                                {{ $t('app.' + item[0]) }}
                            </a>
                        </template>
                        <span class="resources">
                            {{ map.get(item[0]).toString() }}
                        </span>
                    </el-descriptions-item>
                </el-descriptions>
                <div v-if="installData.key === 'openresty'" class="mt-5">
                    <el-checkbox v-model="forceDelete" label="true">{{ $t('app.forceDelete') }}</el-checkbox>
                    <ErrPrompt :title="$t('app.openrestyDeleteHelper')" />
                </div>
            </el-col>
        </el-row>

        <template #footer v-if="forceDelete">
            <span class="dialog-footer">
                <el-button @click="open = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button type="primary" @click="onConfirm">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-dialog>
</template>
<script lang="ts" setup>
import { App } from '@/api/interface/app';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { DeleteRuntime } from '@/api/modules/runtime';
const router = useRouter();

interface CheckProps {
    items: App.AppInstallResource[];
    installID: Number;
    key: string;
}
const installData = ref<CheckProps>({
    items: [],
    installID: 0,
    key: '',
});
const open = ref(false);
const map = new Map();
const forceDelete = ref(false);
const em = defineEmits(['close']);

const acceptParams = (props: CheckProps) => {
    map.clear();
    forceDelete.value = false;
    installData.value.installID = props.installID;
    installData.value.key = props.key;
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
    if (key === 'website') {
        router.push({ name: 'Website' });
    }
};

const onConfirm = () => {
    ElMessageBox.confirm(
        i18n.global.t('app.operatorHelper', [i18n.global.t('app.delete')]),
        i18n.global.t('app.delete'),
        {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'info',
        },
    ).then(() => {
        const params = {
            id: installData.value.installID.valueOf(),
            forceDelete: true,
        };
        DeleteRuntime(params).then(() => {
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            open.value = false;
            em('close', open);
        });
    });
};

defineExpose({
    acceptParams,
});
</script>

<style scoped>
.resources {
    word-break: break-all;
}
</style>
