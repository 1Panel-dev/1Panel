<template>
    <div v-loading="loading">
        <LayoutContent :title="'MySQL ' + $t('menu.database')">
            <template #toolbar>
                <el-row>
                    <el-col :xs="24" :sm="20" :md="20" :lg="20" :xl="20">
                        <el-button type="primary" @click="onOpenDialog('create')">
                            {{ $t('database.create') }}
                        </el-button>
                    </el-col>
                    <el-col :xs="24" :sm="4" :md="4" :lg="4" :xl="4">
                        <div class="search-button">
                            <el-input
                                v-model="searchName"
                                clearable
                                @clear="search()"
                                suffix-icon="Search"
                                @keyup.enter="search()"
                                @change="search()"
                                :placeholder="$t('commons.button.search')"
                            ></el-input>
                        </div>
                    </el-col>
                </el-row>
            </template>
            <template #main>
                <ComplexTable :pagination-config="paginationConfig" @sort-change="search" @search="search" :data="data">
                    <el-table-column :label="$t('commons.table.name')" prop="name" sortable />
                    <el-table-column :label="$t('commons.login.username')" prop="address" />
                    <el-table-column :label="$t('commons.login.username')" prop="username" />
                    <el-table-column :label="$t('commons.login.password')" prop="password">
                        <template #default="{ row }">
                            <div>
                                <span style="float: left; line-height: 25px" v-if="!row.showPassword">***********</span>
                                <div style="cursor: pointer; float: left" v-if="!row.showPassword">
                                    <el-icon
                                        style="margin-left: 5px; margin-top: 3px"
                                        @click="row.showPassword = true"
                                        :size="16"
                                    >
                                        <View />
                                    </el-icon>
                                </div>
                                <span style="float: left" v-if="row.showPassword">{{ row.password }}</span>
                                <div style="cursor: pointer; float: left" v-if="row.showPassword">
                                    <el-icon
                                        style="margin-left: 5px; margin-top: 3px"
                                        @click="row.showPassword = false"
                                        :size="16"
                                    >
                                        <Hide />
                                    </el-icon>
                                </div>
                                <div style="cursor: pointer; float: left">
                                    <el-icon style="margin-left: 5px; margin-top: 3px" :size="16" @click="onCopy(row)">
                                        <DocumentCopy />
                                    </el-icon>
                                </div>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column
                        prop="createdAt"
                        :label="$t('commons.table.date')"
                        :formatter="dateFormat"
                        show-overflow-tooltip
                    />
                </ComplexTable>
            </template>
        </LayoutContent>

        <OperateDialog ref="dialogRef" @search="search" />
    </div>
</template>

<script lang="ts" setup>
import { dateFormat } from '@/utils/util';
import { reactive, ref } from 'vue';
import { searchRemoteDBs } from '@/api/modules/database';
import i18n from '@/lang';
import { MsgError, MsgSuccess } from '@/utils/message';
import useClipboard from 'vue-clipboard3';
import { Database } from '@/api/interface/database';
const { toClipboard } = useClipboard();

const loading = ref(false);

const dialogRef = ref();

const data = ref();
const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 10,
    total: 0,
});
const searchName = ref();

const search = async (column?: any) => {
    let params = {
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
        info: searchName.value,
        orderBy: column?.order ? column.prop : 'created_at',
        order: column?.order ? column.order : 'null',
    };
    const res = await searchRemoteDBs(params);
    data.value = res.data.items || [];
    paginationConfig.total = res.data.total;
};

const onOpenDialog = async (
    title: string,
    rowData: Partial<Database.RemoteDBInfo> = {
        name: '',
        type: 'Mysql',
        address: '',
        port: 3306,
        username: '',
        password: '',
        description: '',
    },
) => {
    let params = {
        title,
        rowData: { ...rowData },
    };
    dialogRef.value!.acceptParams(params);
};

const onCopy = async (row: any) => {
    try {
        await toClipboard(row.password);
        MsgSuccess(i18n.global.t('commons.msg.copySuccess'));
    } catch (e) {
        MsgError(i18n.global.t('commons.msg.copyfailed'));
    }
};

// const onDelete = async (row: Database.MysqlDBInfo) => {
//     const res = await deleteCheckMysqlDB(row.id);
//     deleteRef.value.acceptParams({ id: row.id, name: row.name });
// };

// const buttons = [
//     {
//         label: i18n.global.t('commons.button.delete'),
//         click: (row: Database.MysqlDBInfo) => {
//             onDelete(row);
//         },
//     },
// ];
</script>
