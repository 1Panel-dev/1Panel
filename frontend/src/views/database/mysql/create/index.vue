<template>
    <DrawerPro v-model="createVisible" :header="$t('commons.button.create')" @close="handleClose" size="normal">
        <el-form ref="formRef" label-position="top" :model="form" :rules="rules" v-loading="loading">
            <el-form-item :label="$t('commons.table.name')" prop="name">
                <el-input clearable v-model.trim="form.name" @input="syncUsername" />
            </el-form-item>
            <el-form-item :label="$t('database.format')" prop="format">
                <el-select filterable v-model="form.format" @change="loadCollations()">
                    <el-option
                        v-for="item of formatOptions"
                        :key="item.format"
                        :label="item.format"
                        :value="item.format"
                    />
                </el-select>
            </el-form-item>
            <el-form-item :label="$t('database.collation')" prop="collation">
                <el-select filterable v-model="form.collation" clearable>
                    <el-option v-for="item of collationOptions" :key="item" :label="item" :value="item" />
                </el-select>
                <span class="input-help">{{ $t('database.collationHelper', [form.format]) }}</span>
            </el-form-item>
            <el-form-item :label="$t('database.userBind')" prop="userMode">
                <el-radio-group v-model="form.userMode" @change="changeUserMode">
                    <el-radio-button value="none">
                        {{ $t('database.noUserBind') }}
                    </el-radio-button>
                    <el-radio-button value="select" :disabled="!users.length">
                        {{ $t('commons.button.select') }}
                    </el-radio-button>
                    <el-radio-button value="create">
                        {{ $t('commons.button.create') }}
                    </el-radio-button>
                </el-radio-group>
            </el-form-item>
            <el-form-item v-if="form.userMode === 'select'" :label="$t('commons.login.username')" prop="userKey">
                <el-select
                    v-model="form.userKey"
                    filterable
                    popper-class="mysql-create-user-select-popper"
                    @change="syncUser"
                >
                    <el-option
                        v-for="item in users"
                        :key="item.username + '@' + item.host"
                        class="user-select-option"
                        :label="item.username + ' - ' + permissionLabel(item.host)"
                        :value="item.username + '@' + item.host"
                    >
                        <div class="user-option-row">
                            <div class="user-option">
                                <span class="user-option-title">
                                    {{ item.username }} - {{ permissionLabel(item.host) }}
                                </span>
                                <span v-if="item.description" class="user-option-description">
                                    {{ item.description }}
                                </span>
                            </div>
                            <el-popover
                                :disabled="userDatabases(item).length === 0"
                                placement="right"
                                trigger="hover"
                                :width="220"
                            >
                                <template #reference>
                                    <el-button class="user-bound-button" size="small" @click.stop>
                                        {{ $t('commons.status.bound') }} {{ userDatabases(item).length }}
                                    </el-button>
                                </template>
                                <div class="user-bound-list">
                                    <el-tag
                                        v-for="database in userDatabases(item)"
                                        :key="item.username + item.host + database"
                                        size="small"
                                    >
                                        {{ database }}
                                    </el-tag>
                                </div>
                            </el-popover>
                        </div>
                    </el-option>
                </el-select>
            </el-form-item>
            <el-form-item v-if="form.userMode === 'create'" :label="$t('commons.login.username')" prop="username">
                <el-input clearable v-model.trim="form.username" />
            </el-form-item>
            <el-form-item v-if="form.userMode === 'create'" :label="$t('commons.login.password')" prop="password">
                <el-input type="password" clearable show-password v-model.trim="form.password">
                    <template #append>
                        <el-button @click="random">{{ $t('commons.button.random') }}</el-button>
                    </template>
                </el-input>
                <span class="input-help">{{ $t('commons.rule.illegalChar') }}</span>
            </el-form-item>
            <el-form-item v-if="form.userMode === 'create'" :label="$t('database.permission')" prop="permission">
                <el-select v-model="form.permission">
                    <el-option value="%" :label="$t('database.permissionAll')" />
                    <el-option
                        v-if="form.from !== 'local'"
                        value="localhost"
                        :label="$t('terminal.localhost') + '(localhost)'"
                    />
                    <el-option value="ip" :label="$t('database.permissionForIP')" />
                </el-select>
                <span v-if="form.from !== 'local'" class="input-help">
                    {{ $t('database.localhostHelper') }}
                </span>
            </el-form-item>
            <el-form-item v-if="form.userMode === 'create' && form.permission === 'ip'" prop="permissionIPs">
                <el-input clearable :rows="3" type="textarea" v-model="form.permissionIPs" />
                <span class="input-help">{{ $t('database.remoteHelper') }}</span>
            </el-form-item>
            <el-form-item :label="$t('commons.table.type')" prop="database">
                <el-tag>{{ form.database + ' [' + form.type + ']' }}</el-tag>
            </el-form-item>
            <el-form-item :label="$t('commons.table.description')" prop="description">
                <el-input type="textarea" clearable v-model="form.description" />
            </el-form-item>
        </el-form>
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
import { reactive, ref } from 'vue';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElForm } from 'element-plus';
import { Database } from '@/api/interface/database';
import {
    addMysqlDB,
    grantMysqlUser,
    loadFormatCollations,
    searchMysqlGrants,
    searchMysqlUsers,
} from '@/api/modules/database';
import { MsgSuccess } from '@/utils/message';
import { getRandomStr } from '@/utils/id';
const loading = ref();
const createVisible = ref(false);
const formatOptions = ref();
const collationOptions = ref();
const users = ref<Database.MysqlUser[]>([]);
const grants = ref<Database.MysqlGrant[]>([]);
const form = reactive({
    name: '',
    from: 'local',
    type: '',
    database: '',
    format: '',
    collation: '',
    createUser: false,
    userMode: 'none',
    userKey: '',
    username: '',
    host: '%',
    password: '',
    permission: '',
    permissionIPs: '',
    description: '',
});
const rules = reactive({
    name: [Rules.requiredInput, Rules.dbName],
    format: [Rules.requiredSelect],
    userMode: [Rules.requiredSelect],
    userKey: [Rules.requiredSelect],
    username: [Rules.requiredInput, Rules.name],
    password: [Rules.requiredInput, Rules.noSpace, Rules.illegal],
    permission: [Rules.requiredSelect],
    permissionIPs: [Rules.requiredInput, Rules.noSpace, Rules.illegal],
});

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
    form.format = 'utf8mb4';
    form.collation = '';
    form.createUser = false;
    form.userMode = 'none';
    form.userKey = '';
    form.username = '';
    form.host = '%';
    form.permission = '%';
    form.permissionIPs = '';
    form.description = '';
    random();
    loadOptions();
    await Promise.all([loadUsers(), loadGrants()]);
    if (users.value.length) {
        form.userMode = 'select';
        const user = users.value[0];
        form.userKey = `${user.username}@${user.host}`;
        syncUser();
    }
    createVisible.value = true;
};
const handleClose = () => {
    createVisible.value = false;
};

const loadOptions = async () => {
    const defaultOptions = [{ format: 'utf8mb4' }, { format: 'utf8mb3' }, { format: 'gbk' }, { format: 'big5' }];
    await loadFormatCollations(form.database).then((res) => {
        formatOptions.value = res.data || defaultOptions;
        loadCollations();
    });
};

const loadCollations = async () => {
    collationOptions.value = formatOptions.value?.find((item) => item.format === form.format)?.collations || [];
};

const loadUsers = async () => {
    const res = await searchMysqlUsers({ database: form.database });
    users.value = res.data || [];
};

const loadGrants = async () => {
    const res = await searchMysqlGrants({ database: form.database });
    grants.value = res.data || [];
};

const permissionLabel = (host: string) => {
    return host === '%' ? i18n.global.t('database.permissionAll') : host;
};

const userDatabases = (user: Database.MysqlUser) => {
    return Array.from(
        new Set(
            grants.value
                .filter((item) => item.username === user.username && item.host === user.host)
                .map((item) => item.database),
        ),
    );
};

const syncUser = () => {
    const item = users.value.find((item) => `${item.username}@${item.host}` === form.userKey);
    if (!item) return;
    form.username = item.username;
    form.host = item.host;
};

const changeUserMode = () => {
    if (form.userMode === 'none') {
        form.userKey = '';
        form.username = '';
        form.host = '%';
        return;
    }
    if (form.userMode === 'select') {
        const user = users.value?.[0];
        form.userKey = user ? `${user.username}@${user.host}` : '';
        syncUser();
        return;
    }
    form.userKey = '';
    form.host = '%';
    form.permission = '%';
    form.permissionIPs = '';
    syncUsername();
    random();
};

const syncUsername = async () => {
    if (form.userMode === 'create') {
        form.username = form.name;
    }
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
        try {
            const params = {
                name: form.name,
                from: form.from,
                database: form.database,
                format: form.format,
                collation: form.collation,
                username: form.username,
                password: form.password,
                permission: form.permission,
                description: form.description,
            };
            if (form.userMode !== 'create') {
                params.username = '';
                params.password = '';
                params.permission = '%';
            }
            if (form.userMode === 'create' && form.permission === 'ip') {
                params.permission = form.permissionIPs;
            }
            await addMysqlDB(params);
            if (form.userMode === 'select') {
                await grantMysqlUser({
                    database: form.database,
                    db: form.name,
                    username: form.username,
                    host: form.host,
                });
            }
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            emit('search');
            createVisible.value = false;
        } finally {
            loading.value = false;
        }
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
</style>

<style lang="scss">
.mysql-create-user-select-popper {
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
