<template>
    <div>
        <ComplexTable :pagination-config="paginationConfig" @search="search" :data="data">
            <template #toolbar>
                <el-button type="primary" @click="onCreate">{{ $t('mail.createAlias') }}</el-button>
            </template>
            <el-table-column :label="$t('mail.source')" prop="source" min-width="200" show-overflow-tooltip />
            <el-table-column :label="$t('mail.target')" prop="target" min-width="200" show-overflow-tooltip />
            <fu-table-operations :ellipsis="10" :buttons="buttons" :label="$t('commons.table.operate')" fixed="right" fix />
        </ComplexTable>

        <!-- Create dialog -->
        <DialogPro v-model="createOpen" :title="$t('mail.createAlias')" @close="handleClose">
            <el-row>
                <el-col :span="22" :offset="1">
                    <el-form ref="formRef" label-position="top" :model="form" :rules="rules">
                        <el-form-item :label="$t('mail.source')" prop="source">
                            <el-input v-model.trim="form.source" placeholder="alias">
                                <template #append>@{{ props.domainName }}</template>
                            </el-input>
                        </el-form-item>
                        <el-form-item :label="$t('mail.target')" prop="target">
                            <el-input v-model.trim="form.target" placeholder="user@example.com" />
                        </el-form-item>
                    </el-form>
                </el-col>
            </el-row>
            <template #footer>
                <el-button @click="handleClose" :disabled="formLoading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit(formRef)" :loading="formLoading">{{ $t('commons.button.confirm') }}</el-button>
            </template>
        </DialogPro>

        <OpDialog ref="opRef" @search="search()" />
    </div>
</template>

<script lang="ts" setup>
import { onMounted, reactive, ref } from 'vue';
import { FormInstance } from 'element-plus';
import { searchMailAliases, createMailAlias, deleteMailAlias } from '@/api/modules/mail';
import { Mail } from '@/api/interface/mail';
import { Rules } from '@/global/form-rules';
import { MsgSuccess } from '@/utils/message';
import i18n from '@/lang';

const props = defineProps<{ domainId: number; domainName: string }>();

const data = ref<Mail.AliasRes[]>([]);
const createOpen = ref(false);
const formLoading = ref(false);
const formRef = ref<FormInstance>();
const opRef = ref();

const form = ref({ source: '', target: '' });
const rules = ref({ source: [Rules.requiredInput], target: [Rules.requiredInput] });

const paginationConfig = reactive({
    cacheSizeKey: 'mail-alias-page-size',
    currentPage: 1,
    pageSize: Number(localStorage.getItem('mail-alias-page-size')) || 20,
    total: 0,
});

const buttons = [
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: Mail.AliasRes) => onDelete(row),
    },
];

const search = async () => {
    const res = await searchMailAliases({
        domainID: props.domainId,
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
    });
    data.value = res.data.items || [];
    paginationConfig.total = res.data.total;
};

const onCreate = () => {
    form.value = { source: '', target: '' };
    createOpen.value = true;
};

const onDelete = (row: Mail.AliasRes) => {
    opRef.value.acceptParams({
        title: i18n.global.t('mail.deleteAlias'),
        names: [row.source],
        msg: i18n.global.t('mail.deleteAliasHelper'),
        api: deleteMailAlias,
        params: { id: row.id },
    });
};

const handleClose = () => { createOpen.value = false; formRef.value?.resetFields(); };

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) return;
        formLoading.value = true;
        createMailAlias({ domainID: props.domainId, source: form.value.source, target: form.value.target })
            .then(() => {
                MsgSuccess(i18n.global.t('commons.msg.createSuccess'));
                handleClose();
                search();
            })
            .finally(() => { formLoading.value = false; });
    });
};

onMounted(() => search());
</script>
