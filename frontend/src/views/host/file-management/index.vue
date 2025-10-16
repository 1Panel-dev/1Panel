<template>
    <div @dragover="handleDragover" @drop="handleDrop" @dragleave="handleDragleave">
        <el-tabs
            type="card"
            class="file-tabs"
            v-model="editableTabsKey"
            @tab-change="changeTab"
            @tab-remove="removeTab"
        >
            <el-tab-pane
                closable
                v-for="item in editableTabs"
                :key="item.id"
                :label="item.name == '' ? $t('file.root') : item.name"
                :name="item.id"
            >
                <div>
                    <div class="flex sm:flex-row flex-col justify-start gap-y-2 items-center gap-x-4" ref="toolRef">
                        <div class="flex-shrink-0 flex sm:w-min w-full items-center justify-start">
                            <el-tooltip :content="$t('file.back')" placement="top">
                                <el-button icon="Back" @click="back" circle />
                            </el-tooltip>
                            <el-tooltip :content="$t('file.right')" placement="top">
                                <el-button icon="Right" @click="right" circle />
                            </el-tooltip>
                            <el-tooltip :content="$t('file.top')" placement="top">
                                <el-button icon="Top" @click="top" circle :disabled="paths.length == 0" />
                            </el-tooltip>
                            <el-tooltip :content="$t('commons.button.refresh')" placement="top">
                                <el-button icon="Refresh" circle @click="search" />
                            </el-tooltip>
                            <el-tooltip
                                :content="req.showHidden ? $t('file.noShowHide') : $t('file.showHide')"
                                placement="top"
                            >
                                <el-button
                                    class="btn"
                                    circle
                                    :type="req.showHidden ? '' : 'primary'"
                                    :icon="req.showHidden ? View : Hide"
                                    @click="viewHideFile"
                                />
                            </el-tooltip>
                        </div>
                        <div class="flex-1 sm:w-min w-full hidden sm:block" :ref="(el) => setPathRef(item.id, el)">
                            <div
                                v-show="!searchableStatus"
                                @click="searchableStatus = true"
                                class="address-bar shadow-md rounded-md px-4 py-2 flex items-center flex-grow"
                            >
                                <div ref="breadCrumbRef" class="flex items-center address-url">
                                    <span class="root mr-2">
                                        <el-link @click.stop="jump('/')">
                                            <el-icon :size="20"><HomeFilled /></el-icon>
                                        </el-link>
                                    </span>
                                    <span
                                        v-for="(path, index) in paths"
                                        :key="path.url"
                                        class="inline-flex items-center"
                                    >
                                        <span class="mr-2 arrow">></span>
                                        <template v-if="index === 0 && hidePaths.length > 0">
                                            <el-dropdown>
                                                <span
                                                    class="path-segment cursor-pointer mr-2 pathname focus:outline-none focus-visible:outline-none"
                                                >
                                                    ..
                                                </span>
                                                <template #dropdown>
                                                    <el-dropdown-menu>
                                                        <el-dropdown-item
                                                            v-for="hidePath in hidePaths"
                                                            :key="hidePath.url"
                                                            @click.stop="jump(hidePath.url)"
                                                        >
                                                            <el-tooltip
                                                                class="box-item"
                                                                effect="dark"
                                                                :content="hidePath.name"
                                                                placement="bottom"
                                                            >
                                                                {{
                                                                    hidePath.name.length > 25
                                                                        ? hidePath.name.substring(0, 22) + '...'
                                                                        : hidePath.name
                                                                }}
                                                            </el-tooltip>
                                                        </el-dropdown-item>
                                                    </el-dropdown-menu>
                                                </template>
                                            </el-dropdown>
                                            <span class="mr-2 arrow">></span>
                                            <el-tooltip
                                                class="box-item"
                                                effect="dark"
                                                :content="path.name"
                                                placement="bottom"
                                            >
                                                <el-link
                                                    class="path-segment cursor-pointer mr-2 pathname"
                                                    @click.stop="jump(path.url)"
                                                >
                                                    {{
                                                        path.name.length > 25
                                                            ? path.name.substring(0, 22) + '...'
                                                            : path.name
                                                    }}
                                                </el-link>
                                            </el-tooltip>
                                        </template>
                                        <template v-else>
                                            <el-tooltip
                                                class="box-item"
                                                effect="dark"
                                                :content="path.name"
                                                placement="bottom"
                                            >
                                                <el-link
                                                    class="path-segment cursor-pointer mr-2 pathname"
                                                    @click.stop="jump(path.url)"
                                                >
                                                    {{
                                                        path.name.length > 25
                                                            ? path.name.substring(0, 22) + '...'
                                                            : path.name
                                                    }}
                                                </el-link>
                                            </el-tooltip>
                                        </template>
                                    </span>
                                </div>
                            </div>
                            <el-input
                                :ref="(el) => setSearchableInputRef(item.id, el)"
                                v-show="searchableStatus"
                                v-model="searchablePath"
                                @blur="searchableInputBlur"
                                class="px-4 py-2 border rounded-md shadow-md"
                                @keyup.enter="
                                    jump(searchablePath);
                                    searchableStatus = false;
                                "
                            />
                        </div>
                        <div class="flex-1 sm:w-min w-full sm:hidden block">
                            <div class="address-bar shadow-md rounded-md px-4 py-2 flex items-center flex-grow">
                                <div class="flex items-center address-url">
                                    <span class="root mr-2">
                                        <el-link @click.stop="jump('/')">
                                            <el-icon :size="20"><HomeFilled /></el-icon>
                                        </el-link>
                                    </span>
                                    <span
                                        v-for="(path, index) in paths"
                                        :key="path.url"
                                        class="inline-flex items-center"
                                    >
                                        <span class="mr-2 arrow">></span>
                                        <template v-if="index === 0 && hidePaths.length > 0">
                                            <el-dropdown>
                                                <span
                                                    class="path-segment cursor-pointer mr-2 pathname focus:outline-none focus-visible:outline-none"
                                                >
                                                    ..
                                                </span>
                                                <template #dropdown>
                                                    <el-dropdown-menu>
                                                        <el-dropdown-item
                                                            v-for="hidePath in hidePaths"
                                                            :key="hidePath.url"
                                                            @click.stop="jump(hidePath.url)"
                                                        >
                                                            <el-tooltip
                                                                class="box-item"
                                                                effect="dark"
                                                                :content="hidePath.name"
                                                                placement="bottom"
                                                            >
                                                                {{
                                                                    hidePath.name.length > 25
                                                                        ? hidePath.name.substring(0, 22) + '...'
                                                                        : hidePath.name
                                                                }}
                                                            </el-tooltip>
                                                        </el-dropdown-item>
                                                    </el-dropdown-menu>
                                                </template>
                                            </el-dropdown>
                                            <span class="mr-2 arrow">></span>
                                            <el-tooltip
                                                class="box-item"
                                                effect="dark"
                                                :content="path.name"
                                                placement="bottom"
                                            >
                                                <el-link
                                                    class="path-segment cursor-pointer mr-2 pathname"
                                                    @click.stop="jump(path.url)"
                                                >
                                                    {{
                                                        path.name.length > 25
                                                            ? path.name.substring(0, 22) + '...'
                                                            : path.name
                                                    }}
                                                </el-link>
                                            </el-tooltip>
                                        </template>
                                    </span>
                                </div>
                            </div>
                        </div>
                    </div>
                    <LayoutContent :title="$t('menu.files')" v-loading="loading">
                        <template #prompt>
                            <el-alert type="info" :closable="false">
                                <template #title>
                                    <span class="input-help whitespace-break-spaces">
                                        {{ $t('file.fileHelper') }}
                                    </span>
                                </template>
                            </el-alert>
                        </template>
                        <template #leftToolBar>
                            <div ref="leftWrapper" class="flex items-center gap-2 flex-wrap">
                                <el-dropdown @command="handleCreate" class="mr-2.5">
                                    <el-button type="primary">
                                        {{ $t('commons.button.create') }}
                                        <el-icon><arrow-down /></el-icon>
                                    </el-button>
                                    <template #dropdown>
                                        <el-dropdown-menu>
                                            <el-dropdown-item command="dir">
                                                <svg-icon iconName="p-file-folder"></svg-icon>
                                                {{ $t('file.dir') }}
                                            </el-dropdown-item>
                                            <el-dropdown-item command="file">
                                                <svg-icon iconName="p-file-normal"></svg-icon>
                                                {{ $t('menu.files') }}
                                            </el-dropdown-item>
                                        </el-dropdown-menu>
                                    </template>
                                </el-dropdown>
                                <el-dropdown class="mr-2.5">
                                    <el-button>
                                        {{ $t('commons.button.upload') }}/{{ $t('commons.button.download') }}
                                        <el-icon><arrow-down /></el-icon>
                                    </el-button>
                                    <template #dropdown>
                                        <el-dropdown-menu>
                                            <el-dropdown-item @click="openUpload">
                                                <el-icon><ElUpload /></el-icon>
                                                {{ $t('commons.button.upload') }}
                                            </el-dropdown-item>
                                            <el-dropdown-item @click="openWget">
                                                <el-icon><ElDownload /></el-icon>
                                                {{ $t('file.remoteFile') }}
                                            </el-dropdown-item>
                                        </el-dropdown-menu>
                                    </template>
                                </el-dropdown>
                                <el-button-group class="sm:!inline-block !flex flex-wrap gap-y-2">
                                    <el-button class="btn" @click="openRecycleBin">
                                        {{ $t('file.recycleBin') }}
                                    </el-button>
                                    <el-button class="btn" @click="toTerminal">
                                        {{ $t('menu.terminal') }}
                                    </el-button>
                                    <el-popover
                                        placement="bottom"
                                        :width="250"
                                        trigger="hover"
                                        @before-enter="getFavorites"
                                    >
                                        <template #reference>
                                            <el-button @click="openFavorite">
                                                {{ $t('file.favorite') }}
                                            </el-button>
                                        </template>
                                        <div class="favorite-item">
                                            <el-table :data="favorites">
                                                <el-table-column prop="name">
                                                    <template #default="{ row }">
                                                        <div class="flex justify-between items-center group">
                                                            <el-tooltip
                                                                class="box-item"
                                                                effect="dark"
                                                                :content="row.path"
                                                                placement="top"
                                                            >
                                                                <span
                                                                    class="table-link text-ellipsis"
                                                                    @click="toFavorite(row)"
                                                                    type="primary"
                                                                >
                                                                    <svg-icon
                                                                        v-if="row.isDir"
                                                                        className="table-icon"
                                                                        iconName="p-file-folder"
                                                                    ></svg-icon>
                                                                    <svg-icon
                                                                        v-else
                                                                        className="table-icon"
                                                                        iconName="p-file-normal"
                                                                    ></svg-icon>
                                                                    {{ row.name }}
                                                                </span>
                                                            </el-tooltip>
                                                            <el-icon
                                                                class="hidden group-hover:block cursor-pointer"
                                                                v-if="!row.isDir"
                                                                @click="jump(row.path)"
                                                            >
                                                                <FolderOpened />
                                                            </el-icon>
                                                        </div>
                                                    </template>
                                                </el-table-column>
                                            </el-table>
                                        </div>
                                    </el-popover>
                                    <el-button class="btn" @click="calculateSize(req.path)" :loading="disableBtn">
                                        {{ $t('file.calculate') }}
                                    </el-button>
                                    <template v-if="hostMount.length == 1">
                                        <el-button class="btn" @click.stop="jump(hostMount[0]?.path)">
                                            {{ hostMount[0]?.path }} ({{ $t('file.root') }})
                                            {{ formatFileSize(hostMount[0]?.free) }}
                                        </el-button>
                                    </template>
                                    <template v-else>
                                        <el-dropdown class="mr-2.5">
                                            <el-button class="btn">
                                                {{ hostMount[0]?.path }} ({{ $t('file.root') }})
                                                {{ formatFileSize(hostMount[0]?.free) }}
                                            </el-button>
                                            <template #dropdown>
                                                <el-dropdown-menu>
                                                    <template v-for="(mount, index) in hostMount" :key="mount.path">
                                                        <el-dropdown-item
                                                            v-if="index == 0"
                                                            @click.stop="jump(mount.path)"
                                                        >
                                                            {{ mount.path }} ({{ $t('file.root') }})
                                                            {{ formatFileSize(mount.free) }}
                                                        </el-dropdown-item>
                                                        <el-dropdown-item
                                                            v-if="index != 0"
                                                            @click.stop="jump(mount.path)"
                                                        >
                                                            {{ mount.path }} ({{ $t('home.mount') }})
                                                            {{ formatFileSize(mount.free) }}
                                                        </el-dropdown-item>
                                                    </template>
                                                </el-dropdown-menu>
                                            </template>
                                        </el-dropdown>
                                    </template>
                                </el-button-group>

                                <el-badge :value="processCount" class="btn" v-if="processCount > 0">
                                    <el-button class="btn" @click="openProcess">
                                        {{ $t('file.wgetTask') }}
                                    </el-button>
                                </el-badge>
                            </div>
                        </template>
                        <template #rightToolBar>
                            <div :ref="(el) => setBtnWrapperRef(item.id, el)" class="flex items-center gap-2 flex-wrap">
                                <div class="flex items-center gap-2 flex-wrap">
                                    <template v-if="visibleButtons.length == 0">
                                        <el-dropdown v-if="moreButtons.length">
                                            <el-button>
                                                {{ $t('tabs.more') }}
                                                <i class="el-icon-arrow-down el-icon--right" />
                                            </el-button>
                                            <template #dropdown>
                                                <el-dropdown-menu>
                                                    <el-dropdown-item
                                                        v-for="btn in moreButtons"
                                                        :key="btn.label"
                                                        @click="btn.action"
                                                        :disabled="selects.length === 0"
                                                    >
                                                        {{ $t(btn.label) }}
                                                    </el-dropdown-item>
                                                </el-dropdown-menu>
                                            </template>
                                        </el-dropdown>
                                    </template>
                                    <template v-if="visibleButtons.length > 0">
                                        <el-button-group class="flex items-center">
                                            <template v-for="btn in visibleButtons" :key="btn.label">
                                                <el-button plain @click="btn.action" :disabled="selects.length === 0">
                                                    {{ $t(btn.label) }}
                                                </el-button>
                                            </template>

                                            <el-dropdown v-if="moreButtons.length">
                                                <el-button>
                                                    {{ $t('tabs.more') }}
                                                    <i class="el-icon-arrow-down el-icon--right" />
                                                </el-button>
                                                <template #dropdown>
                                                    <el-dropdown-menu>
                                                        <el-dropdown-item
                                                            v-for="btn in moreButtons"
                                                            :key="btn.label"
                                                            @click="btn.action"
                                                            :disabled="selects.length === 0"
                                                        >
                                                            {{ $t(btn.label) }}
                                                        </el-dropdown-item>
                                                    </el-dropdown-menu>
                                                </template>
                                            </el-dropdown>
                                        </el-button-group>
                                    </template>
                                </div>
                                <el-button-group class="copy-button" v-if="moveOpen">
                                    <el-tooltip
                                        class="box-item"
                                        effect="dark"
                                        :content="$t('file.paste')"
                                        placement="bottom"
                                    >
                                        <el-button plain @click="openPaste">
                                            {{ $t('file.paste') }}({{ fileMove.count }})
                                        </el-button>
                                    </el-tooltip>
                                    <el-tooltip
                                        class="box-item"
                                        effect="dark"
                                        :content="$t('commons.button.cancel')"
                                        placement="bottom"
                                    >
                                        <el-button plain class="close" icon="Close" @click="closeMove"></el-button>
                                    </el-tooltip>
                                </el-button-group>
                                <div class="w-80">
                                    <el-input
                                        v-model="req.search"
                                        clearable
                                        @clear="search()"
                                        @keydown.enter="search()"
                                        :placeholder="$t('file.search')"
                                    >
                                        <template #prepend>
                                            <el-checkbox v-model="req.containSub">
                                                {{ $t('file.sub') }}
                                            </el-checkbox>
                                        </template>
                                        <template #append>
                                            <el-button icon="Search" @click="search" round />
                                        </template>
                                    </el-input>
                                </div>
                            </div>
                        </template>
                        <template #main>
                            <ComplexTable
                                :pagination-config="paginationConfig"
                                v-model:selects="selects"
                                :ref="(el) => setTableRef(item.id, el)"
                                :data="data"
                                @search="search"
                                @sort-change="changeSort"
                                @cell-mouse-enter="showFavorite"
                                @cell-mouse-leave="hideFavorite"
                                :heightDiff="heightDiff"
                                :right-buttons="buttons"
                            >
                                <el-table-column type="selection" width="30" />
                                <el-table-column
                                    :label="$t('commons.table.name')"
                                    min-width="250"
                                    fix
                                    show-overflow-tooltip
                                    :sortable="'custom'"
                                    prop="name"
                                >
                                    <template #default="{ row }">
                                        <div class="file-row">
                                            <div>
                                                <svg-icon
                                                    v-if="row.isDir"
                                                    className="table-icon"
                                                    iconName="p-file-folder"
                                                ></svg-icon>
                                                <svg-icon
                                                    v-else
                                                    className="table-icon"
                                                    :iconName="getIconName(row.extension)"
                                                ></svg-icon>
                                            </div>
                                            <div class="file-name">
                                                <el-input
                                                    v-if="fileRename.oldName === row.name && isEdit"
                                                    v-model.trim="fileRename.newName"
                                                    :ref="(el) => setRenameRef(item.id, el)"
                                                    :autofocus="isEdit"
                                                    class="table-link table-input"
                                                    placeholder="file name"
                                                    @keydown.enter="handleRename(row)"
                                                    @blur="onRenameBlur($event, row)"
                                                />
                                                <span v-else class="table-link" @click="open(row)" type="primary">
                                                    {{ row.name }}
                                                </span>
                                                <span v-if="row.isSymlink">-> {{ row.linkPath }}</span>
                                            </div>
                                            <div>
                                                <el-button
                                                    v-if="row.favoriteID > 0"
                                                    link
                                                    type="warning"
                                                    size="large"
                                                    icon="StarFilled"
                                                    @click="remove(row.favoriteID)"
                                                ></el-button>
                                                <div v-else>
                                                    <el-button
                                                        v-if="hoveredRowPath === row.path"
                                                        link
                                                        icon="Star"
                                                        @click="addToFavorite(row)"
                                                    ></el-button>
                                                </div>
                                            </div>
                                        </div>
                                    </template>
                                </el-table-column>
                                <el-table-column :label="$t('file.mode')" prop="mode" min-width="110">
                                    <template #default="{ row }">
                                        <el-link underline="never" @click="openMode(row)">{{ row.mode }}</el-link>
                                    </template>
                                </el-table-column>
                                <el-table-column
                                    :label="$t('commons.table.user')"
                                    prop="user"
                                    show-overflow-tooltip
                                    min-width="90"
                                >
                                    <template #default="{ row }">
                                        <el-link underline="never" @click="openChown(row)">
                                            {{ row.user ? row.user : '-' }} ({{ row.uid }})
                                        </el-link>
                                    </template>
                                </el-table-column>
                                <el-table-column :label="$t('file.group')" prop="group" show-overflow-tooltip>
                                    <template #default="{ row }">
                                        <el-link underline="never" @click="openChown(row)">
                                            {{ row.group ? row.group : '-' }} ({{ row.gid }})
                                        </el-link>
                                    </template>
                                </el-table-column>
                                <el-table-column
                                    :label="$t('file.size')"
                                    prop="size"
                                    min-width="100"
                                    :sortable="'custom'"
                                >
                                    <template #default="{ row }">
                                        <el-button
                                            type="primary"
                                            link
                                            small
                                            :loading="row.btnLoading"
                                            @click="row.isDir ? getDirSize(row.path) : getFileSize(row.path)"
                                        >
                                            <span v-if="row.isDir">
                                                <span v-if="row.dirSize === undefined">
                                                    {{ $t('file.calculate') }}
                                                </span>
                                                <span v-else>{{ formatFileSize(row.dirSize) }}</span>
                                            </span>
                                            <span v-else>
                                                {{ formatFileSize(row.size) }}
                                            </span>
                                        </el-button>
                                    </template>
                                </el-table-column>
                                <el-table-column
                                    :label="$t('file.updateTime')"
                                    prop="modTime"
                                    width="180"
                                    :formatter="dateFormat"
                                    show-overflow-tooltip
                                    :sortable="'custom'"
                                ></el-table-column>
                                <fu-table-operations
                                    :ellipsis="mobile ? 0 : 2"
                                    :buttons="buttons"
                                    :label="$t('commons.table.operate')"
                                    :min-width="mobile ? 'auto' : 200"
                                    :fixed="mobile ? false : 'right'"
                                    width="270"
                                    fix
                                />
                                <template #paginationLeft>
                                    <div class="flex justify-start items-center">
                                        <el-text small>
                                            {{ $t('file.fileDirNum', [dirNum, fileNum]) }}
                                        </el-text>
                                        <el-text small>
                                            {{ $t('file.currentDir') + $t('file.size') + ' ' }}
                                        </el-text>
                                        <el-button type="primary" link small :loading="calculateBtn">
                                            <span v-if="dirTotalSize == -1" @click="getDirTotalSize(req.path)">
                                                {{ $t('file.calculate') }}
                                            </span>
                                            <span v-else>
                                                {{ formatFileSize(dirTotalSize) }}
                                            </span>
                                        </el-button>
                                    </div>
                                </template>
                            </ComplexTable>
                        </template>
                    </LayoutContent>
                </div>
            </el-tab-pane>
            <el-tab-pane :closable="false" :disabled="editableTabs.length > 6">
                <template #label>
                    <el-icon @click="addTab()"><Plus /></el-icon>
                </template>
            </el-tab-pane>
        </el-tabs>

        <CreateFile ref="createRef" @close="search" />
        <ChangeRole ref="roleRef" @close="search" />
        <Compress ref="compressRef" @close="search" />
        <Decompress ref="deCompressRef" @close="search" />
        <CodeEditor ref="codeEditorRef" @close="search" />
        <FileRename ref="renameRef" @close="search" />
        <Upload ref="uploadRef" @close="search" />
        <Wget ref="wgetRef" @close="closeWget" />
        <Move ref="moveRef" @close="closeMovePage" @loading="onLoading" />
        <Download ref="downloadRef" @close="search" />
        <Process ref="processRef" @close="closeProcess" />
        <Owner ref="chownRef" @close="search"></Owner>
        <Detail ref="detailRef" />
        <DeleteFile ref="deleteRef" @close="search" />
        <RecycleBin ref="recycleBinRef" @close="search" />
        <Favorite ref="favoriteRef" @close="search" @jump="jump" @toFavorite="toFavorite" />
        <BatchRole ref="batchRoleRef" @close="search" />
        <VscodeOpenDialog ref="dialogVscodeOpenRef" />
        <Preview ref="previewRef" />
        <TerminalDialog ref="dialogTerminalRef" />
        <Convert ref="convertRef" @close="search" />
    </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, reactive, ref } from 'vue';
import {
    addFavorite,
    computeDepthDirSize,
    computeDirSize,
    fileWgetKeys,
    getFileContent,
    getFilesList,
    removeFavorite,
    renameRile,
    searchFavorite,
    searchHostMount,
} from '@/api/modules/files';
import {
    computeSize,
    copyText,
    dateFormat,
    downloadFile,
    getFileType,
    getIcon,
    getRandomStr,
    isConvertible,
} from '@/utils/util';
import { File } from '@/api/interface/file';
import { Languages, Mimetypes } from '@/global/mimetype';
import { useRouter } from 'vue-router';
import { MsgSuccess, MsgWarning } from '@/utils/message';
import { useMultipleSearchable } from './hooks/searchable';
import { ResultData } from '@/api/interface';
import { GlobalStore } from '@/store';
import { Download as ElDownload, Upload as ElUpload, View, Hide } from '@element-plus/icons-vue';

import i18n from '@/lang';
import CreateFile from './create/index.vue';
import ChangeRole from './change-role/index.vue';
import Compress from './compress/index.vue';
import Decompress from './decompress/index.vue';
import Upload from './upload/index.vue';
import FileRename from './rename/index.vue';
import CodeEditor from './code-editor/index.vue';
import Wget from './wget/index.vue';
import Move from './move/index.vue';
import Download from './download/index.vue';
import Owner from './chown/index.vue';
import DeleteFile from './delete/index.vue';
import Process from './process/index.vue';
import Detail from './detail/index.vue';
import RecycleBin from './recycle-bin/index.vue';
import Favorite from './favorite/index.vue';
import BatchRole from './batch-role/index.vue';
import Preview from './preview/index.vue';
import VscodeOpenDialog from '@/components/vscode-open/index.vue';
import Convert from './convert/index.vue';
import { debounce } from 'lodash-es';
import TerminalDialog from './terminal/index.vue';
import { Dashboard } from '@/api/interface/dashboard';
import { CompressExtension, CompressType } from '@/enums/files';
import type { TabPaneName } from 'element-plus';
import { getComponentInfo } from '@/api/modules/host';
import { routerToNameWithQuery } from '@/utils/router';

const globalStore = GlobalStore();

interface FilePaths {
    url: string;
    name: string;
}

const router = useRouter();
const data = ref();
const tableRefs = ref<Record<string, any>>({});
const heightDiff = ref(365);

const setTableRef = (key: string, el: any) => {
    if (el) {
        tableRefs.value[key] = el;
    }
};
const getCurrentTable = () => tableRefs.value[editableTabsKey.value];

let selects = ref<any>([]);

const initData = () => ({
    path: '/',
    expand: true,
    showHidden: localStorage.getItem('show-hidden') === 'true',
    page: 1,
    pageSize: 100,
    search: '',
    containSub: false,
    sortBy: 'name',
    sortOrder: 'ascending',
});
let req = reactive(initData());
let loading = ref(false);
const paths = ref<FilePaths[]>([]);
const hidePaths = ref<FilePaths[]>([]);
let pathWidth = ref(0);
const history: string[] = [];
let pointer = -1;

const fileCreate = reactive({ path: '/', isDir: false, mode: 0o755 });
const fileCompress = reactive({ files: [''], name: '', dst: '', operate: 'compress' });
const fileDeCompress = reactive({ path: '', name: '', dst: '', type: '' });
const fileEdit = reactive({ content: '', path: '', name: '', language: 'plaintext', extension: '' });
const filePreview = reactive({ path: '', name: '', extension: '', fileType: '', imageFiles: [], currentNode: '' });
const codeReq = reactive({ path: '', expand: false, page: 1, pageSize: 100, isDetail: false });
const fileUpload = reactive({ path: '' });
const fileRename = reactive({ path: '', oldName: '', newName: '' });
const fileWget = reactive({ path: '' });
const fileMove = reactive({ oldPaths: [''], allNames: [''], type: '', path: '', name: '', count: 0, isDir: false });
const fileConvert = reactive<{
    outputPath: string;
    files: File.ConvertFile[];
}>({
    outputPath: '',
    files: [
        {
            type: '',
            inputFile: '',
            extension: '',
            path: '',
            outputFormat: '',
        },
    ],
});
const ffmpegExist = ref(false);

const createRef = ref();
const roleRef = ref();
const detailRef = ref();
const compressRef = ref();
const deCompressRef = ref();
const codeEditorRef = ref();
const renameRef = ref();
const uploadRef = ref();
const wgetRef = ref();
const moveRef = ref();
const downloadRef = ref();
const toolRef = ref();
const breadCrumbRef = ref();
const chownRef = ref();
const moveOpen = ref(false);
const deleteRef = ref();
const recycleBinRef = ref();
const favoriteRef = ref();
const hoveredRowPath = ref(null);
const favorites = ref([]);
const batchRoleRef = ref();
const dialogVscodeOpenRef = ref();
const previewRef = ref();
const processRef = ref();
const hostMount = ref<Dashboard.DiskInfo[]>([]);
let resizeObserver: ResizeObserver;
const dirTotalSize = ref(-1);
const disableBtn = ref(false);
const calculateBtn = ref(false);
const dirNum = ref(0);
const fileNum = ref(0);
const imageFiles = ref([]);
const isEdit = ref(false);
const convertRef = ref();

const renameRefs = ref<Record<string, any>>({});

const setRenameRef = (key: string, el: any) => {
    if (el) {
        renameRefs.value[key] = el;
    }
};
const getCurrentRename = () => renameRefs.value[editableTabsKey.value];

const pathRefs = ref<Record<string, any>>({});

const setPathRef = (key: string, el: any) => {
    if (el) {
        pathRefs.value[key] = el;
    }
};
const getCurrentPath = () => pathRefs.value[editableTabsKey.value];

const { searchableStatus, searchablePath, setSearchableInputRef, searchableInputBlur } = useMultipleSearchable(paths);

const paginationConfig = reactive({
    cacheSizeKey: 'file-page-size',
    currentPage: 1,
    pageSize: Number(localStorage.getItem('file-page-size')) || 100,
    total: 0,
});

const mobile = computed(() => {
    return globalStore.isMobile();
});

const search = async () => {
    dirTotalSize.value = -1;
    await getWgetProcess();
    loading.value = true;
    if (req.search != '') {
        req.sortBy = 'name';
        req.sortOrder = 'ascending';
        getCurrentTable().clearSort();
    }
    req.page = paginationConfig.currentPage;
    req.pageSize = paginationConfig.pageSize;
    await getFilesList(req)
        .then((res) => {
            handleSearchResult(res);
        })
        .finally(() => {
            loading.value = false;
        });
};

const searchFile = async () => {
    loading.value = true;
    dirTotalSize.value = -1;
    try {
        return await getFilesList(req);
    } finally {
        loading.value = false;
    }
};

const handleSearchResult = (res: ResultData<File.File>) => {
    data.value = res.data.items || [];
    paginationConfig.total = res.data.itemTotal;
    dirNum.value = data.value.filter((item) => item.isDir).length;
    fileNum.value = data.value.filter((item) => !item.isDir).length;
    req.path = res.data.path;
};

const viewHideFile = async () => {
    req.showHidden = !req.showHidden;
    localStorage.setItem('show-hidden', req.showHidden ? 'true' : 'false');
    let searchResult = await searchFile();
    handleSearchResult(searchResult);
};

const open = async (row: File.File) => {
    hideRightMenu();
    calculateBtn.value = false;
    disableBtn.value = false;
    if (row.isDir) {
        if (row.name.indexOf('.1panel_clash') > -1) {
            MsgWarning(i18n.global.t('file.clashOpenAlert'));
            return;
        }
        const name = row.name;
        if (req.path.endsWith('/')) {
            req.path = req.path + name;
        } else {
            req.path = req.path + '/' + name;
        }
        paths.value.push({
            url: req.path,
            name: name,
        });
        await jump(req.path);
    } else {
        openView(row);
    }
};

const copyDir = (row: File.File) => {
    if (row?.path) {
        copyText(row?.path);
    }
};

const leftWrapper = ref<HTMLElement | null>(null);
const btnWrapperRefs = ref<Record<string, any>>({});

const setBtnWrapperRef = (key: string, el: any) => {
    if (el) {
        btnWrapperRefs.value[key] = el;
    }
};
const getCurrentBtnWrapper = () => btnWrapperRefs.value[editableTabsKey.value];

const toolButtons = ref([
    {
        label: 'commons.button.copy',
        action: () => openMove('copy'),
    },
    {
        label: 'file.move',
        action: () => openMove('cut'),
    },
    {
        label: 'file.compress',
        action: () => openCompress(selects.value),
    },
    {
        label: 'file.role',
        action: () => openBatchRole(selects.value),
    },
    {
        label: 'commons.button.delete',
        action: () => batchDelFiles(),
    },
]);

const visibleButtons = ref([...toolButtons.value]);
const moreButtons = ref([]);

const updateButtons = async () => {
    await nextTick();
    if (!getCurrentBtnWrapper()) return;
    const pathWidth = toolRef.value.offsetWidth;
    const leftWidth = leftWrapper.value.offsetWidth;
    let num = Math.floor((pathWidth - leftWidth - 450) / 100);
    if (num < 0) {
        visibleButtons.value = toolButtons.value;
        moreButtons.value = [];
    } else {
        visibleButtons.value = toolButtons.value.slice(0, num);
        moreButtons.value = toolButtons.value.slice(num);
    }
};

const handlePath = () => {
    nextTick(function () {
        let breadCrumbWidth = breadCrumbRef.value?.offsetWidth;
        let pathWidth = toolRef.value?.offsetWidth;
        if (pathWidth - breadCrumbWidth < 50 && paths.value.length > 1) {
            const removed = paths.value.shift();
            if (removed) hidePaths.value.push(removed);
            handlePath();
        }
    });
};

const resizeHandler = debounce(() => {
    resetPaths();
    handlePath();
}, 100);

const btnResizeHandler = debounce(() => {
    updateButtons();
}, 100);

const observeResize = () => {
    const el = getCurrentPath() as any;
    if (!el) return;
    let resizeObserver = new ResizeObserver(() => {
        resizeHandler();
    });

    const ele = getCurrentBtnWrapper() as any;
    if (!ele) return;
    resizeObserver = new ResizeObserver(() => {
        btnResizeHandler();
    });
    resizeObserver.observe(el);
    resizeObserver.observe(ele);
};

function watchTitleHeight() {
    const el = document.querySelector<HTMLElement>('.content-container__title');
    if (el) {
        let titleHeight = el.offsetHeight < 40 ? 40 : 80;
        heightDiff.value = 325 + titleHeight;
    }
}

watchTitleHeight();

window.addEventListener('resize', watchTitleHeight);

const resetPaths = () => {
    paths.value = [...hidePaths.value, ...paths.value];
    hidePaths.value = [];
};

const right = () => {
    if (pointer < history.length - 1) {
        pointer++;
        let url = history[pointer];
        backForwardJump(url);
    }
};

const back = () => {
    if (pointer > 0) {
        pointer--;
        let url = history[pointer];
        backForwardJump(url);
    }
};

const top = () => {
    if (paths.value.length > 0) {
        let url = '/';
        if (paths.value.length >= 2) {
            url = paths.value[paths.value.length - 2].url;
        }
        jump(url);
    }
};

const jump = async (url: string) => {
    hideRightMenu();
    history.splice(pointer + 1);
    history.push(url);
    pointer = history.length - 1;

    const { path: oldUrl, pageSize: oldPageSize } = req;
    Object.assign(req, initData(), { path: url, containSub: false, search: '', pageSize: oldPageSize });
    let searchResult = await searchFile();
    if (!searchResult.data.path) {
        req.path = oldUrl;
        globalStore.setLastFilePath(req.path);
        MsgWarning(i18n.global.t('commons.res.notFound'));
        return;
    }
    req.path = searchResult.data.path;
    globalStore.setLastFilePath(req.path);
    handleSearchResult(searchResult);
    getPaths(req.path);
    updateTab(req.path);
    await nextTick(function () {
        handlePath();
    });
};

const backForwardJump = async (url: string) => {
    const oldPageSize = req.pageSize;
    Object.assign(req, initData());
    req.path = url;
    req.containSub = false;
    req.search = '';
    req.pageSize = oldPageSize;
    let searchResult = await searchFile();
    handleSearchResult(searchResult);
    getPaths(req.path);
    updateTab(req.path);
    await nextTick(function () {
        handlePath();
    });
};

const getPaths = (reqPath: string) => {
    const pathArray = reqPath.split('/');
    paths.value = [];
    hidePaths.value = [];
    let base = '/';
    for (const p of pathArray) {
        if (p != '') {
            if (base.endsWith('/')) {
                base = base + p;
            } else {
                base = base + '/' + p;
            }
            paths.value.push({
                url: base,
                name: p,
            });
        }
    }
};

const handleCreate = (command: string) => {
    fileCreate.path = req.path;
    fileCreate.isDir = command === 'dir';
    createRef.value.acceptParams(fileCreate);
};

const delFile = async (row: File.File | null) => {
    deleteRef.value.acceptParams([row]);
};

const batchDelFiles = () => {
    deleteRef.value.acceptParams(selects.value);
};

const formatFileSize = (size: number) => {
    return computeSize(size);
};

const getFileSize = async (path: string) => {
    codeReq.path = path;
    codeReq.expand = true;
    codeReq.isDetail = true;
    updateByPath(path, { btnLoading: true });
    try {
        const res = await getFileContent(codeReq);
        updateByPath(path, { dirSize: res.data.size });
    } finally {
        updateByPath(path, { btnLoading: false });
    }
};

const getDirSize = async (path: string) => {
    const req = {
        path: path,
    };
    updateByPath(path, { btnLoading: true });
    try {
        const res = await computeDirSize(req);
        updateByPath(path, { dirSize: res.data.size });
    } finally {
        updateByPath(path, { btnLoading: false });
    }
};

const updateByPath = (path: string, patch: Partial<(typeof data.value)[0]>) => {
    data.value = data.value.map((item) => (item.path === path ? { ...item, ...patch } : item));
};

const getDirTotalSize = async (path: string) => {
    const req = {
        path: path,
    };
    calculateBtn.value = true;
    const res = await computeDirSize(req);
    dirTotalSize.value = res.data.size;
    calculateBtn.value = false;
};

const calculateSize = (path: string) => {
    const req = { path };
    disableBtn.value = true;
    setTimeout(async () => {
        try {
            const res = await computeDepthDirSize(req);
            const sizeMap = new Map(res.data.map((dir) => [dir.path, dir.size]));
            data.value.forEach((item) => {
                if (sizeMap.has(item.path)) {
                    item.dirSize = sizeMap.get(item.path)!;
                }
            });
        } catch (err) {
            console.error('Error computing dir size:', err);
        } finally {
            disableBtn.value = false;
        }
    }, 0);
};

const getIconName = (extension: string) => {
    return getIcon(extension);
};

const openMode = (item: File.File) => {
    roleRef.value.acceptParams(item);
};

const openChown = (item: File.File) => {
    chownRef.value.acceptParams(item);
};

const openCompress = (items: File.File[]) => {
    const paths = [];
    for (const item of items) {
        paths.push(item.path);
    }
    fileCompress.files = paths;
    if (paths.length === 1) {
        fileCompress.name = items[0].name;
    } else {
        fileCompress.name = getRandomStr(6);
    }
    fileCompress.dst = req.path;

    compressRef.value.acceptParams(fileCompress);
};

const openDeCompress = (item: File.File) => {
    if (Mimetypes.get(item.mimeType) == undefined) {
        MsgWarning(i18n.global.t('file.canNotDeCompress'));
        return;
    }
    fileDeCompress.type = Mimetypes.get(item.mimeType);
    if (CompressExtension[Mimetypes.get(item.mimeType)] != item.extension) {
        fileDeCompress.type = getEnumKeyByValue(item.extension);
    }
    if (item.name.endsWith('.tar.gz') || item.name.endsWith('.tgz')) {
        fileDeCompress.type = CompressType.TarGz;
    }

    fileDeCompress.name = item.name;
    fileDeCompress.path = item.path;
    fileDeCompress.dst = req.path;

    deCompressRef.value.acceptParams(fileDeCompress);
};

function getEnumKeyByValue(value: string): keyof typeof CompressExtension | undefined {
    return (Object.keys(CompressExtension) as Array<keyof typeof CompressExtension>).find(
        (k) => CompressExtension[k] === value,
    );
}

const openView = (item: File.File) => {
    const fileType = getFileType(item.extension);
    if (fileType === 'image') {
        imageFiles.value = data.value
            .filter((item) => !item.isDir)
            .filter((item) => getFileType(item.extension) == 'image')
            .map((item) => (item.isSymlink ? item.linkPath : item.path));
    }

    const previewTypes = ['image', 'video', 'audio', 'word', 'excel'];
    if (previewTypes.includes(fileType)) {
        return openPreview(item, fileType);
    }

    const actionMap = {
        compress: openDeCompress,
        text: () => openCodeEditor(item.path, item.extension),
    };

    const path = item.isSymlink ? item.linkPath : item.path;
    return actionMap[fileType] ? actionMap[fileType](item) : openCodeEditor(path, item.extension);
};

const openPreview = (item: File.File, fileType: string) => {
    if (item.mode.toString() == '-' && item.user == '-' && item.group == '-') {
        MsgWarning(i18n.global.t('file.fileCanNotRead'));
        return;
    }
    filePreview.path = item.isSymlink ? item.linkPath : item.path;
    filePreview.name = item.name;
    filePreview.extension = item.extension;
    filePreview.fileType = fileType;
    filePreview.imageFiles = imageFiles.value;
    filePreview.currentNode = globalStore.currentNode;

    previewRef.value.acceptParams(filePreview);
};

const openCodeEditor = (path: string, extension: string) => {
    codeReq.path = path;
    codeReq.expand = true;

    if (extension != '') {
        Languages.forEach((language) => {
            const ext = extension.substring(1);
            if (language.value.indexOf(ext) > -1) {
                fileEdit.language = language.label;
            }
        });
    }

    getFileContent(codeReq)
        .then((res) => {
            fileEdit.content = res.data.content;
            fileEdit.path = res.data.path;
            fileEdit.name = res.data.name;
            fileEdit.extension = res.data.extension;

            codeEditorRef.value.acceptParams(fileEdit);
        })
        .catch(() => {});
};

const openUpload = () => {
    fileUpload.path = req.path;
    uploadRef.value.acceptParams(fileUpload);
};

const openWget = () => {
    fileWget.path = req.path;
    wgetRef.value.acceptParams(fileWget);
};

const openBatchRole = (items: File.File[]) => {
    batchRoleRef.value.acceptParams({ files: items });
};

const closeWget = (submit: Boolean) => {
    search();
    if (submit) {
        openProcess();
    }
};

const closeMovePage = (submit: Boolean) => {
    if (submit) {
        search();
        closeMove();
    }
};

const openProcess = () => {
    processRef.value.acceptParams();
};

const closeProcess = () => {
    search();
    getWgetProcess();
    setTimeout(() => {
        getWgetProcess();
    }, 1000);
};

const processCount = ref(0);
const getWgetProcess = async () => {
    processCount.value = 0;
    try {
        const res = await fileWgetKeys();
        if (res.data && res.data.keys.length > 0) {
            processCount.value = res.data.keys.length;
        }
    } catch (error) {}
};

const openRename = (item: File.File) => {
    fileRename.path = req.path;
    fileRename.oldName = item.name;
    fileRename.newName = item.name;
    isEdit.value = true;
    nextTick(() => {
        getCurrentRename().focus();
    });
    hideRightMenu();
};

const onRenameBlur = (e: FocusEvent, row: File.File) => {
    const related = e.relatedTarget as HTMLElement | null;
    if (
        related &&
        (related.closest('.fu-table-more-button') || related.closest('.fu-table-more-button .el-dropdown__item'))
    ) {
        setTimeout(() => {
            getCurrentRename()?.focus();
        }, 0);
        return;
    }
    handleRename(row);
};

const handleRename = async (row: File.File): Promise<void> => {
    if (fileRename.newName === fileRename.oldName) {
        isEdit.value = false;
        fileRename.oldName = '';
        return;
    }
    const addItem: File.FileRename = {
        oldName: getPath(fileRename.path, fileRename.oldName),
        newName: getPath(fileRename.path, fileRename.newName),
    };
    loading.value = true;
    try {
        await renameRile(addItem);
        MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
        row.name = fileRename.newName;
        row.path = getPath(req.path, fileRename.newName);
    } catch (error) {
        console.error(error);
    } finally {
        loading.value = false;
        isEdit.value = false;
        fileRename.oldName = '';
    }
};

const getPath = (path: string, name: string) => {
    return path + '/' + name;
};

const openMove = (type: string) => {
    fileMove.type = type;
    fileMove.name = '';
    fileMove.allNames = [];
    fileMove.isDir = false;
    const oldPaths = [];
    for (const s of selects.value) {
        oldPaths.push(s['path']);
    }
    fileMove.count = selects.value.length;
    fileMove.oldPaths = oldPaths;
    if (selects.value.length == 1) {
        fileMove.name = selects.value[0].name;
        fileMove.isDir = selects.value[0].isDir;
    } else {
        const allNames = [];
        for (const s of selects.value) {
            allNames.push(s['name']);
        }
        fileMove.allNames = allNames;
    }
    moveOpen.value = true;
    if (type === 'cut') {
        MsgSuccess(i18n.global.t('file.moveSuccess') + '! ' + i18n.global.t('file.pasteMsg'));
    } else {
        MsgSuccess(i18n.global.t('file.copySuccess') + '! ' + i18n.global.t('file.pasteMsg'));
    }
};

const openMoveBtn = (type: string, item: File.File) => {
    selects.value = [];
    selects.value.push(item);
    openMove(type);
};

const closeMove = () => {
    selects.value = [];
    getCurrentTable().clearSelects();
    hideRightMenu();
    fileMove.oldPaths = [];
    fileMove.name = '';
    fileMove.count = 0;
    fileMove.isDir = false;
    moveOpen.value = false;
};

const openPaste = () => {
    fileMove.path = req.path;
    moveRef.value.acceptParams(fileMove);
};

function onLoading(isLoading: boolean) {
    loading.value = isLoading;
}

const openDownload = (file: File.File) => {
    downloadFile(file.path, globalStore.currentNode);
};

const openDetail = (row: File.File) => {
    detailRef.value.acceptParams({ path: row.path });
};

const openRecycleBin = () => {
    recycleBinRef.value.acceptParams();
};

const openFavorite = () => {
    favoriteRef.value.acceptParams();
};

const changeSort = ({ prop, order }) => {
    req.sortBy = prop;
    req.sortOrder = order;
    req.search = '';
    req.page = 1;
    req.pageSize = paginationConfig.pageSize;
    req.containSub = false;
    search();
};

const showFavorite = (row: File.File) => {
    hoveredRowPath.value = row.path;
};

const hideFavorite = () => {
    hoveredRowPath.value = null;
};

const addToFavorite = async (row: File.File) => {
    try {
        await addFavorite(row.path);
        await search();
    } catch (error) {}
};

const remove = async (id: number) => {
    ElMessageBox.confirm(i18n.global.t('file.removeFavorite'), i18n.global.t('commons.msg.remove'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    }).then(async () => {
        try {
            await removeFavorite(id);
            await search();
        } catch (error) {}
    });
};

const getFavorites = async () => {
    try {
        const res = await searchFavorite(req);
        favorites.value = res.data.items;
    } catch (error) {}
};

const toFavorite = (row: File.Favorite) => {
    if (row.isDir) {
        jump(row.path);
    } else {
        let file = {} as File.File;
        file.path = row.path;
        file.extension = '.' + row.name.split('.').pop();
        openView(file);
    }
};

const dialogTerminalRef = ref();
const toTerminal = () => {
    dialogTerminalRef.value!.acceptParams({ cwd: req.path, command: '/bin/sh' });
};

const openWithVSCode = (row: File.File) => {
    dialogVscodeOpenRef.value.acceptParams({ path: row.path + (row.isDir ? '' : ':1:1') });
};

const buttons = [
    {
        label: i18n.global.t('commons.button.open'),
        click: open,
    },
    {
        label: i18n.global.t('commons.button.download'),
        click: (row: File.File) => {
            openDownload(row);
        },
        disabled: (row: File.File) => {
            return row?.isDir;
        },
    },
    {
        label: i18n.global.t('commons.button.copy'),
        click: (row: File.File) => openMoveBtn('copy', row),
    },
    {
        label: i18n.global.t('file.move'),
        click: (row: File.File) => openMoveBtn('cut', row),
    },
    {
        label: i18n.global.t('file.paste'),
        click: openPaste,
        disabled: () => {
            return !moveOpen.value;
        },
    },
    {
        label: i18n.global.t('file.compress'),
        click: (row: File.File) => {
            openCompress([row]);
        },
    },
    {
        label: i18n.global.t('file.deCompress'),
        click: openDeCompress,
        disabled: (row: File.File) => {
            return !isDecompressFile(row);
        },
    },
    {
        label: i18n.global.t('file.editPermissions'),
        click: (row: File.File) => {
            openBatchRole([row]);
        },
    },
    {
        label: i18n.global.t('file.rename'),
        click: openRename,
    },
    {
        label: i18n.global.t('commons.button.delete'),
        disabled: (row: File.File) => {
            return row.name == '.1panel_clash';
        },
        click: delFile,
        divided: true,
    },
    {
        label: i18n.global.t('file.copyDir'),
        click: copyDir,
    },
    {
        label: i18n.global.t('file.addFavorite'),
        click: (row: File.File) => {
            if (row?.favoriteID > 0) {
                remove(row?.favoriteID);
            } else {
                addToFavorite(row);
            }
        },
    },
    {
        label: i18n.global.t('file.convert'),
        click: (row: File.File) => {
            openConvert(row);
        },
        disabled: (row: File.File) => {
            return row?.isDir || !isConvertible(row?.extension, row?.mimeType);
        },
    },
    {
        label: i18n.global.t('file.openWithVscode'),
        click: openWithVSCode,
    },
    {
        label: i18n.global.t('file.info'),
        click: openDetail,
        divided: true,
    },
];

const openConvert = (item: File.File) => {
    if (!ffmpegExist.value) {
        ElMessageBox.confirm(i18n.global.t('cronjob.library.noSuchApp', ['FFmpeg']), i18n.global.t('file.convert'), {
            confirmButtonText: i18n.global.t('app.toInstall'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
        }).then(() => {
            routerToNameWithQuery('Library', { t: Date.now(), uncached: 'true' });
        });
        return;
    } else {
        if (!isConvertible(item.extension, item.mimeType)) {
            MsgWarning(i18n.global.t('file.fileCanNotConvert'));
            return;
        }
        const fileType = getFileType(item.extension);
        fileConvert.outputPath = req.path;
        fileConvert.files = [
            {
                type: fileType,
                path: req.path,
                extension: item.extension,
                inputFile: item.name,
                outputFormat: item.extension.slice(1),
            },
        ];

        convertRef.value.acceptParams(fileConvert);
    }
};

const isDecompressFile = (row: File.File) => {
    if (row.isDir) {
        return false;
    }
    if (getFileType(row.extension) === 'compress') {
        return true;
    }
    if (row.mimeType == 'application/octet-stream') {
        return false;
    } else {
        return Mimetypes.get(row.mimeType) != undefined;
    }
};

const getHostMount = async () => {
    try {
        const res = await searchHostMount();
        hostMount.value = res.data;
    } catch (error) {
        console.error('Error fetching host mount:', error);
    }
};

const handleDrop = async (event: DragEvent) => {
    event.preventDefault();
    fileUpload.path = req.path;
    if (!uploadRef.value?.open) {
        await uploadRef.value?.handleDrop(event);
    }
    uploadRef.value.acceptParams(fileUpload);
};

const handleDragover = (event: DragEvent) => {
    event.preventDefault();
};

const handleDragleave = (event: { preventDefault: () => void }) => {
    event.preventDefault();
};

function hideRightMenu() {
    getCurrentTable().closeRightClick();
}

onMounted(() => {
    initShowHidden();
    initTabsAndPaths();
    getHostMount();
    initHistory();
    checkFFmpeg();
    nextTick(function () {
        handlePath();
        observeResize();
    });
});

function initShowHidden() {
    const showHidden = localStorage.getItem('show-hidden');
    if (showHidden === null) {
        localStorage.setItem('show-hidden', 'true');
        req.showHidden = true;
    } else {
        req.showHidden = showHidden === 'true';
    }
}

function initTabsAndPaths() {
    initTabs();
    let path = getInitialPath();
    req.path = path;
    getPaths(path);
    editableTabsValue.value = path;
    updateTab(path);
    paths.value = buildPaths(path);
    pathWidth.value = getCurrentPath()?.offsetWidth;
}

function buildPaths(path: string) {
    const segments = path.split('/').filter(Boolean);
    let url = '';
    return segments.map((segment) => {
        url += '/' + segment;
        return { url, name: segment };
    });
}

function initHistory() {
    search();
    history.push(req.path);
    pointer = history.length - 1;
}

function getInitialPath(): string {
    const routePath = router.currentRoute.value.query.path;
    if (routePath) {
        const p = String(routePath);
        globalStore.setLastFilePath(p);
        return p;
    } else if (globalStore.lastFilePath && globalStore.lastFilePath !== '') {
        return globalStore.lastFilePath;
    }
    return '/';
}

const editableTabsKey = ref('');
const editableTabsValue = ref('');
const editableTabsName = ref('');
const editableTabs = ref([
    { id: '1', name: 'opt', path: '/opt' },
    { id: '2', name: 'home', path: '/home' },
]);

function initTabs() {
    const savedTabs = localStorage.getItem('editableTabs');
    if (savedTabs) {
        editableTabs.value = JSON.parse(savedTabs);
    }

    const savedTabsKey = localStorage.getItem('editableTabsKey');
    if (savedTabsKey) {
        editableTabsKey.value = savedTabsKey;
        const tab = editableTabs.value.find((t) => t.id === savedTabsKey);
        if (tab) {
            editableTabsValue.value = tab.path;
            editableTabsName.value = tab.name;
        } else {
            setFirstTab();
        }
    } else {
        setFirstTab();
    }
}

function setFirstTab() {
    if (editableTabs.value.length > 0) {
        const first = editableTabs.value[0];
        editableTabsKey.value = first.id;
        editableTabsValue.value = first.path;
        editableTabsName.value = first.name;
    }
}

watch(
    [editableTabs, editableTabsKey],
    ([newTabs, newKey]) => {
        localStorage.setItem('editableTabs', JSON.stringify(newTabs));
        localStorage.setItem('editableTabsKey', newKey);
    },
    { deep: true },
);

function getLastPath(path: string): string {
    if (!path) return '';
    const parts = path.split('/').filter(Boolean);
    return parts.length ? parts[parts.length - 1] : '';
}

function updateTab(newPath?: string) {
    const tab = editableTabs.value.find((t) => t.id === editableTabsKey.value);
    if (tab) {
        tab.path = newPath;
        tab.name = getLastPath(newPath);
    }
}

const addTab = () => {
    if (editableTabs.value.length >= 6) {
        MsgWarning(i18n.global.t('file.notCanTab'));
        return;
    }
    const usedIds = editableTabs.value.map((t) => Number(t.id));
    let newId = null;
    for (let i = 1; i <= 6; i++) {
        if (!usedIds.includes(i)) {
            newId = i;
            break;
        }
    }
    if (newId === null) {
        MsgWarning(i18n.global.t('file.notCanTab'));
        return;
    }
    editableTabs.value.push({
        id: String(newId),
        name: 'opt',
        path: '/opt',
    });
    editableTabsKey.value = String(newId);
    changeTab(String(newId));
};

const changeTab = (targetPath: TabPaneName) => {
    if (targetPath === 99) {
        return;
    }
    editableTabsKey.value = targetPath.toString();
    const current = editableTabs.value.find((tab) => tab.id === editableTabsKey.value);
    editableTabsName.value = current ? current.name : '';
    editableTabsValue.value = current ? current.path : '';
    req.path = editableTabsValue.value;
    paths.value = [];
    const segments = editableTabsValue.value.split('/').filter(Boolean);
    let url = '';
    segments.forEach((segment) => {
        url += '/' + segment;
        paths.value.push({
            url,
            name: segment,
        });
    });
    search();
};

const removeTab = (targetId: TabPaneName) => {
    const tabs = editableTabs.value;
    if (tabs.length <= 1) {
        MsgWarning(i18n.global.t('file.keepOneTab'));
        return;
    }
    const target = String(targetId);
    const current = String(editableTabsKey.value);
    const idx = tabs.findIndex((t) => String(t.id) === target);
    if (idx === -1) return;
    let nextActive = current;
    if (current === target) {
        nextActive = tabs[idx + 1]?.id ?? tabs[idx - 1]?.id ?? current;
    }
    editableTabs.value = tabs.filter((t) => String(t.id) !== target);
    editableTabsKey.value = String(nextActive);
    changeTab(String(nextActive));
};

const checkFFmpeg = () => {
    getComponentInfo('ffmpeg', globalStore.currentNode).then((res) => {
        ffmpegExist.value = res.data.exists ?? false;
    });
};

onBeforeUnmount(() => {
    if (resizeObserver) resizeObserver.disconnect();
    window.removeEventListener('resize', watchTitleHeight);
});
</script>

<style scoped lang="scss">
.path {
    display: flex;
    align-items: center;
    border: 1px solid #ebeef5;
    background-color: var(--panel-path-bg);
    height: 30px;
    border-radius: 2px !important;
    &:hover {
        cursor: text;
        box-shadow: var(--el-box-shadow);
    }

    .root {
        vertical-align: middle;
        margin-left: 10px;
    }
    .other {
        vertical-align: middle;
    }
    .split {
        margin-left: 5px;
        margin-right: 5px;
    }
}

.copy-button {
    .close {
        width: 10px;
        .close-icon {
            color: red;
        }
    }
}

.btn-container {
    display: flex;
    justify-content: space-between;
    align-items: center;
    width: 100%;
}

.favorite-item {
    height: 30vh;
    overflow: auto;
}

.file-row {
    display: flex;
    align-items: center;
    width: 100%;
}

.file-name {
    flex-grow: 1;
    margin-left: 1px;
    width: 95%;
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
}
.address-bar {
    border: var(--el-border);
    .arrow {
        color: #726e6e;
    }
}
.search-button {
    width: 20vw;
}
.el-button-group > .el-dropdown > .el-button {
    border-left-color: var(--el-border-color);
}
.table-input {
    --el-input-inner-height: 22px !important;
}
:deep(.el-tabs__nav .el-tabs__item:last-child) {
    border-bottom: 1px solid var(--el-border-color-light) !important;
}

:deep(.file-tabs .el-tabs--card > .el-tabs__header .el-tabs__item.is-active) {
    border-bottom-width: 1px !important;
}
:deep(.file-tabs .el-tabs--card .el-tabs__header .el-tabs__nav) {
    border-bottom: none !important;
}
</style>
