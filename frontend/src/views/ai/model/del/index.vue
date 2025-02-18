<template>
    <div>
        <el-dialog
            v-model="open"
            :title="$t('commons.button.sync')"
            width="30%"
            :close-on-click-modal="false"
            @close="handleClose"
        >
            <div v-loading="loading">
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-alert class="mt-2" :show-icon="true" type="warning" :closable="false">
                            {{ $t('ai_tools.model.ollama_sync') }}
                        </el-alert>
                        <el-checkbox
                            class="mt-2"
                            v-model="checkAll"
                            :indeterminate="isIndeterminate"
                            @change="handleCheckAllChange"
                        >
                            {{ $t('setting.all') }}
                        </el-checkbox>
                        <el-checkbox-group v-model="checkedItems" @change="handleCheckedChange">
                            <el-checkbox
                                v-for="(item, index) in list"
                                :key="index"
                                :label="item.name"
                                :value="item.id"
                            />
                        </el-checkbox-group>
                    </el-col>
                </el-row>
            </div>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="handleClose()" :disabled="loading">
                        {{ $t('commons.button.cancel') }}
                    </el-button>
                    <el-button type="primary" @click="onConfirm" :disabled="loading || checkedItems.length === 0">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-dialog>
    </div>
</template>

<script setup lang="ts">
import { AI } from '@/api/interface/ai';
import { deleteOllamaModel } from '@/api/modules/ai';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { CheckboxValueType } from 'element-plus';
import { onMounted, ref } from 'vue';

defineOptions({ name: 'OpDialog' });

const checkAll = ref(false);
const isIndeterminate = ref(true);
const checkedItems = ref([]);
const list = ref([]);

const loading = ref();
const open = ref();

interface DialogProps {
    list: Array<AI.OllamaModelDropInfo>;
}
const acceptParams = (props: DialogProps): void => {
    list.value = props.list;
    checkAll.value = true;
    handleCheckAllChange(true);
    open.value = true;
};

const emit = defineEmits(['search']);

const handleCheckAllChange = (val: CheckboxValueType) => {
    checkedItems.value = [];
    if (val) {
        for (const item of list.value) {
            checkedItems.value.push(item.id);
        }
    }
    isIndeterminate.value = false;
};
const handleCheckedChange = (value: CheckboxValueType[]) => {
    const checkedCount = value.length;
    checkAll.value = checkedCount === list.value.length;
    isIndeterminate.value = checkedCount > 0 && checkedCount < list.value.length;
};

const onConfirm = async () => {
    loading.value = true;
    await deleteOllamaModel(checkedItems.value, true)
        .then(() => {
            emit('search');
            MsgSuccess(i18n.global.t('commons.msg.deleteSuccess'));
            open.value = false;
            loading.value = false;
        })
        .catch(() => {
            loading.value = false;
        });
};

const handleClose = () => {
    emit('search');
    open.value = false;
};

onMounted(() => {});

defineExpose({
    acceptParams,
});
</script>
