<template>
    <div>
        <el-row>
            <el-col :xs="20" :sm="12" :md="10" :lg="10" :xl="8">
                <el-form :model="form" :rules="rules" ref="formRef">
                    <el-form-item prop="funcs">
                        <el-input
                            type="text"
                            v-model="form.funcs"
                            label="value"
                            :placeholder="$t('php.disableFunctionHelper')"
                        />
                    </el-form-item>
                </el-form>
                <ComplexTable :data="data" v-loading="loading">
                    <template #toolbar>
                        <el-button type="primary" icon="Plus" @click="openCreate(formRef)">
                            {{ $t('commons.button.add') }}
                        </el-button>
                    </template>
                    <el-table-column :label="$t('commons.table.name')" prop="func"></el-table-column>
                    <el-table-column :label="$t('commons.table.operate')">
                        <template #default="{ $index }">
                            <el-button link type="primary" @click="remove($index)">
                                {{ $t('commons.button.delete') }}
                            </el-button>
                        </template>
                    </el-table-column>
                </ComplexTable>
            </el-col>
        </el-row>
        <ConfirmDialog ref="confirmDialogRef" @confirm="submit(false, [''])"></ConfirmDialog>
    </div>
</template>
<script setup lang="ts">
import { GetPHPConfig, UpdatePHPConfig } from '@/api/modules/website';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { FormInstance } from 'element-plus';
import { onMounted, reactive } from 'vue';
import { computed, ref } from 'vue';

const props = defineProps({
    id: {
        type: Number,
        default: 0,
    },
});

const rules = reactive({
    funcs: [Rules.requiredInput, Rules.disabledFunctions],
});

const websiteID = computed(() => {
    return props.id;
});
const formRef = ref();
const loading = ref(false);
const form = ref({
    funcs: '',
});
const data = ref([]);
const confirmDialogRef = ref();

const search = () => {
    loading.value = true;
    GetPHPConfig(websiteID.value)
        .then((res) => {
            const functions = res.data.disableFunctions || [];
            if (functions.length > 0) {
                functions.forEach((value: string) => {
                    data.value.push({
                        func: value,
                    });
                });
            }
        })
        .finally(() => {
            loading.value = false;
        });
};

const openCreate = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        let params = {
            header: i18n.global.t('database.confChange'),
            operationInfo: i18n.global.t('database.restartNowHelper'),
            submitInputInfo: i18n.global.t('database.restartNow'),
        };
        confirmDialogRef.value!.acceptParams(params);
    });
};

const remove = async (index: number) => {
    ElMessageBox.confirm(i18n.global.t('commons.msg.delete'), i18n.global.t('commons.msg.deleteTitle'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    })
        .then(() => {
            const copyList = data.value.concat();
            copyList.splice(index, 1);
            const funcArray: string[] = [];
            copyList.forEach((d) => {
                funcArray.push(d.func);
            });
            submit(true, funcArray);
        })
        .catch(() => {});
};

const submit = async (del: boolean, funcArray: string[]) => {
    let disableFunctions = [];
    if (del) {
        disableFunctions = funcArray;
    } else {
        disableFunctions = form.value.funcs.split(',');
        data.value.forEach((d) => {
            disableFunctions.push(d.func);
        });
    }

    loading.value = true;
    UpdatePHPConfig({ scope: 'disable_functions', id: websiteID.value, disableFunctions: disableFunctions })
        .then(() => {
            form.value.funcs = '';
            data.value = [];
            MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
            search();
        })
        .finally(() => {
            loading.value = false;
        });
};

onMounted(() => {
    search();
});
</script>
