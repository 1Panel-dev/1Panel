<template>
    <DrawerPro v-model="createVisible" :header="$t('commons.button.create')" @close="handleClose" size="normal">
        <div v-loading="loading">
            <el-form ref="formRef" label-position="top" :model="form" :rules="rules">
                <el-form-item :label="$t('commons.table.name')" prop="name">
                    <el-input
                        clearable
                        v-model.trim="form.name"
                        @input="!form.existingUser && (form.username = form.name)"
                    ></el-input>
                </el-form-item>
                <el-form-item :label="$t('database.userAuthorization')">
                    <el-radio-group v-model="form.existingUser" @change="changeUserMode">
                        <el-radio-button :value="true" :disabled="!users.length">
                            {{ $t('commons.button.select') }}
                        </el-radio-button>
                        <el-radio-button :value="false">{{ $t('commons.button.create') }}</el-radio-button>
                    </el-radio-group>
                </el-form-item>
                <el-form-item :label="$t('commons.login.username')" prop="username">
                    <el-select
                        v-if="form.existingUser"
                        v-model="form.username"
                        filterable
                        :fit-input-width="true"
                        popper-class="postgresql-create-user-select-popper"
                    >
                        <el-option
                            v-for="user in users"
                            :key="user.username"
                            :value="user.username"
                            :label="user.username"
                            class="user-select-option"
                        >
                            <div class="user-option-row">
                                <div class="user-option">
                                    <div class="user-option-title">
                                        <el-tooltip :content="user.username" v-bind="TABLE_TOOLTIP_OPTIONS">
                                            <span class="user-name-text">{{ user.username }}</span>
                                        </el-tooltip>
                                        <el-tag v-if="user.superUser" class="superuser-tag" type="warning" size="small">
                                            {{ $t('database.pgSuperUser') }}
                                        </el-tag>
                                    </div>
                                    <span v-if="user.description" class="user-option-description">
                                        {{ user.description }}
                                    </span>
                                </div>
                                <el-tag v-if="user.superUser" size="small">{{ $t('commons.table.all') }}</el-tag>
                                <el-popover
                                    v-else
                                    :disabled="!user.databases.length"
                                    placement="right"
                                    trigger="hover"
                                    :width="220"
                                >
                                    <template #reference>
                                        <el-button class="user-bound-button" size="small" @click.stop>
                                            {{ $t('database.authorizedDatabaseCount', [user.databases.length]) }}
                                        </el-button>
                                    </template>
                                    <div class="user-bound-list">
                                        <el-tag v-for="db in user.databases" :key="db" size="small">{{ db }}</el-tag>
                                    </div>
                                </el-popover>
                            </div>
                        </el-option>
                    </el-select>
                    <el-input v-else clearable v-model.trim="form.username" />
                </el-form-item>
                <el-form-item v-if="!form.existingUser" :label="$t('commons.login.password')" prop="password">
                    <el-input type="password" clearable show-password v-model.trim="form.password">
                        <template #append>
                            <el-button @click="random">{{ $t('commons.button.random') }}</el-button>
                        </template>
                    </el-input>
                    <span class="input-help">{{ $t('commons.rule.illegalChar') }}</span>
                </el-form-item>
                <el-form-item v-if="!form.existingUser" :label="$t('database.permission')" prop="superUser">
                    <el-checkbox v-model="form.superUser">{{ $t('database.pgSuperUser') }}</el-checkbox>
                    <span class="input-help">{{ $t('database.pgAuthorizationHelper') }}</span>
                </el-form-item>

                <el-form-item :label="$t('commons.table.type')" prop="database">
                    <el-tag>{{ form.database + ' [' + form.type + ']' }}</el-tag>
                </el-form-item>

                <el-form-item :label="$t('commons.table.description')" prop="description">
                    <el-input type="textarea" clearable v-model="form.description" />
                </el-form-item>
            </el-form>
        </div>
        <template #footer>
            <span class="dialog-footer">
                <el-button :disabled="loading" @click="createVisible = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button :disabled="loading" type="primary" @click="onSubmit(formRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { computed, reactive, ref } from 'vue';
import { TABLE_TOOLTIP_OPTIONS } from '@/components/table/shared';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElForm } from 'element-plus';
import { addPostgresqlDB, searchPostgresqlUsers } from '@/api/modules/database';
import { MsgSuccess } from '@/utils/message';
import { getRandomStr } from '@/utils/id';
import { Database } from '@/api/interface/database';
const users = ref<Database.PostgresqlUser[]>([]);
const loading = ref();
const createVisible = ref(false);
const form = reactive({
    name: '',
    from: 'local',
    type: '',
    database: '',
    format: '',
    username: '',
    password: '',
    superUser: false,
    existingUser: false,
    description: '',
});
const rules = computed(() => ({
    name: [Rules.requiredInput, Rules.dbName],
    username: form.existingUser ? [Rules.requiredSelect] : [Rules.requiredInput, Rules.name],
    password: [Rules.requiredInput, Rules.databasePassword],
}));

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

interface DialogProps {
    from: string;
    type: string;
    database: string;
}
const acceptParams = async (params: DialogProps): Promise<void> => {
    form.name = '';
    form.from = params.from;
    form.type = params.type;
    form.database = params.database;
    form.format = 'UTF8';
    form.username = '';
    form.superUser = false;
    form.existingUser = false;
    users.value = [];
    form.description = '';
    random();
    createVisible.value = true;
    loading.value = true;
    try {
        users.value = (await searchPostgresqlUsers(params.database)).data || [];
    } catch {
        return;
    } finally {
        loading.value = false;
    }
};
const changeUserMode = () => {
    form.username = form.existingUser ? users.value[0]?.username || '' : form.name;
    form.password = form.existingUser ? '' : getRandomStr(16);
    formRef.value?.clearValidate();
};
const handleClose = () => {
    createVisible.value = false;
};

const random = async () => {
    form.password = getRandomStr(16);
};

const emit = defineEmits<{ (e: 'search'): void }>();
const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;

        loading.value = true;
        await addPostgresqlDB(form)
            .then(() => {
                loading.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                emit('search');
                createVisible.value = false;
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

defineExpose({
    acceptParams,
});
</script>

<style lang="scss" scoped>
:deep(.user-select-option) {
    height: auto;
    min-height: 48px;
}
.user-option-row {
    display: flex;
    align-items: center;
    gap: 8px;
    width: 100%;
}
.user-option {
    display: flex;
    flex-direction: column;
    justify-content: center;
    flex: 1;
    min-width: 0;
    min-height: 40px;
    padding: 4px 0;
    line-height: 18px;
}
.user-option-title {
    display: flex;
    align-items: center;
    gap: 6px;
    min-width: 0;
}
.user-name-text {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}
.user-option-description {
    margin-top: 3px;
    color: var(--el-text-color-secondary);
    font-size: 12px;
    line-height: 16px;
    white-space: normal;
    word-break: break-word;
}
.user-bound-button {
    flex-shrink: 0;
}
.superuser-tag {
    flex-shrink: 0;
}
</style>

<style lang="scss">
.postgresql-create-user-select-popper {
    min-width: 360px;

    .user-select-option {
        height: auto;
        min-height: 48px;
        padding-top: 4px;
        padding-bottom: 4px;
    }
}
.user-bound-list {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
}
</style>
