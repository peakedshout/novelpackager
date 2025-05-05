<script lang="ts">

import {api} from "../tool/api.ts";
import {NewLoadingContext, ProcessError, ProcessResult} from "../tool/tool1.ts";
import {BookInfo, ChapterInfo, VolumeInfo} from "../model/model.ts";

export default {
  data() {
    return {
      showDrawer: false,
      showInfoId: "",
      showSource: "",

      showImageIs: false,
      showImageData: "",
      showImageName: "",

      bookInfo: new BookInfo(),
      showChapterIs: false,
      showChapterVol: "",
      showChapterList: [] as ChapterInfo[],

      lc: NewLoadingContext(),

      cachingMap: new Map<string, string>(),
      showCachingProgress: "",
      showCachingProgressIs: false,
      enableDownloadIs: false,
      enableDownloadShowList: [] as string[],
      downloadShowIs: false,
      downloadVols: [] as boolean[],
    }
  },
  props: {
    getInfoId: {
      type: String,
      required: true,
    },
    getInfoIs: {
      type: Boolean,
      required: true,
    },
    sourceSelect: {
      type: String,
      required: true,
    },
  },
  emits: ['update:getInfoIs', 'update:getInfoId'],
  methods: {
    getBookInfo() {
      this.lc.Loading(async () => {
        const result = await api.GetBookInfo(this.showSource, this.showInfoId)
        const res = ProcessResult<BookInfo>(result)
        if (res) {
          this.bookInfo = res
        }
        this.getSourceProgress()
      })
    },
    getSourceProgress() {
      this.enableDownloadIs = false
      this.showCachingProgressIs = false
      this.lc.Loading(async () => {
        const result = await api.GetProgress(this.showSource)
        const res = ProcessResult<Map<string, string>>(result)
        if (res) {
          this.cachingMap = new Map(Object.entries(res))
          const show = this.cachingMap.get(this.showInfoId)
          if (show) {
            if (show == "100.00%") {
              this.enableDownloadIs = true
              this.showCachingProgressIs = true
              this.showCachingProgress = "try sync cache"
            } else if (show.startsWith("err:")) {
              this.showCachingProgressIs = true
              this.showCachingProgress = show
            } else {
              this.showCachingProgressIs = false
              this.showCachingProgress = "caching progress: " + show
            }
          } else {
            this.showCachingProgressIs = true
            this.showCachingProgress = "try sync cache"
          }
        }
      })
    },
    cachingBook() {
      this.lc.Loading(async () => {
        const result = await api.Caching(this.showSource, this.showInfoId)
        if (ProcessError(result)) {
          this.showCachingProgressIs = false
          this.showCachingProgress = "caching progress: null"
        }
      })
    },
    getEnableDownload() {
      this.downloadVols = []
      this.lc.Loading(async () => {
        const result = await api.GetEnableDownload(this.showSource, this.showInfoId)
        const res = ProcessResult<string[]>(result)
        if (res) {
          this.enableDownloadShowList = res
          this.downloadShowIs = true
          this.downloadVols.fill(false, 0, this.enableDownloadShowList.length)
        }
      })
    },
    downloadBook() {
      this.downloadShowIs = false
      this.enableDownloadShowList = []
      this.lc.Loading(async () => {
        const vols = [] as number [];
        for (let i = 0; i < this.downloadVols.length; i++) {
          if (this.downloadVols[i]) {
            vols.push(i + 1)
          }
        }
        await api.Download(this.showSource, this.showInfoId, vols)
      })
    },
    close() {
      this.$emit("update:getInfoIs", false)
      this.$emit("update:getInfoId", "")
    },
    getBase64Image(base64Data: string) {
      if (!base64Data.startsWith('data:image/')) {
        return `data:image/*;base64,${base64Data}`;
      }
      return base64Data;
    },
    showImage(name: string, base64Data: string) {
      this.showImageName = name
      this.showImageData = base64Data
      this.showImageIs = true
    },
    showChapter(vol: VolumeInfo) {
      if (this.showImageIs) {
        return
      }
      this.showChapterIs = true
      this.showChapterVol = vol.name
      this.showChapterList = vol.chapters
    },
  },
  mounted() {
    this.showDrawer = this.getInfoIs
    this.showInfoId = this.getInfoId
    this.showSource = this.sourceSelect
  },
  watch: {
    getInfoIs(val) {
      this.showDrawer = val
    },
    getInfoId(val) {
      this.showInfoId = val
    },
    sourceSelect(val) {
      this.showSource = val
      this.bookInfo = {} as BookInfo
    },
  }
}


</script>

<template>
  <el-drawer v-model="showDrawer" title="" :with-header="false"
             :before-close="close" @open="getBookInfo"
             body-class="info_bg" size="50%">
    <div v-loading="lc.isLoading" class="backgroundL" v-show="lc.isLoading"/>
    <div v-if="!lc.isLoading" style="height: 100%; width: 100%; display: flex; flex-direction: column">
      <div class="info_header">
        <div style="display: inline-flex; width: 100%">
          <img :src="getBase64Image(bookInfo.cover)" alt="" style="width: auto; height: 100px;"
               @click="showImage(bookInfo.name,bookInfo.cover)">
          <div style="margin-left: 10px; margin-right: 10px; text-align: left; width: 100%; height: 100px">
            <div style="display: inline-flex; justify-content: space-between;width: 100%">
              <el-text size="large">{{ bookInfo.name }}</el-text>
              <div style="display: inline-flex;">
                <el-button type="warning" @click="getBookInfo">
                  🔄
                </el-button>
                <el-tooltip
                    :content="showCachingProgress"
                    placement="top"
                >
                  <el-button type="primary" @click="cachingBook" :disabled="!showCachingProgressIs">
                    ⚡️
                  </el-button>
                </el-tooltip>
                <el-tooltip
                    placement="top"
                >
                  <template #content>
                    <el-text size="small" v-if="enableDownloadIs">Download: {{ bookInfo.name }}</el-text>
                    <el-text size="small" v-else>Need to sync cache: {{ bookInfo.name }}</el-text>
                  </template>
                  <el-button type="success" @click="getEnableDownload" :disabled="!enableDownloadIs">
                    ⬇️
                  </el-button>
                </el-tooltip>
              </div>
            </div>
            <br/>
            <el-text size="default" line-clamp="1">{{ bookInfo.author }}</el-text>
            <br/>
            <br/>
            <div style="display: inline-flex">
              <el-text line-clamp="1" style="flex-wrap: wrap">
                <el-tag v-for="t in bookInfo.metas">{{ t }}</el-tag>
              </el-text>
            </div>
          </div>
        </div>
      </div>
      <div class="info_Des">
        <el-text size="small">{{ bookInfo.description }}</el-text>
      </div>
      <div class="info_Vol">
        <div class="info_VolItem" v-for="v in bookInfo.volumes" @click="showChapter(v)">
          <div style="display: inline-flex">
            <img :src="getBase64Image(v.cover)" alt="" style="width: auto; height: 100px;"
                 @click="showImage(v.name,v.cover)">
            <div style="margin-left: 10px; text-align: left; width: 100%; height: 100px">
              <el-text size="large">{{ v.name }}</el-text>
              <br/>
              <div style="max-height: 80px; overflow-y: auto;">
                <el-text size="small">{{ v.description }}</el-text>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </el-drawer>
  <el-dialog v-model="showImageIs" :title="showImageName" width="50%"
             :before-close="() => {showImageIs = false;showImageName='';}">
    <img :src="getBase64Image(showImageData)" alt="" style="width: 100%; height: 100%">
  </el-dialog>
  <el-dialog v-model="showChapterIs" :title="showChapterVol" width="50%"
             :before-close="() => {showChapterIs = false;showChapterVol='';}">
    <el-table :data="showChapterList" :show-header="false" style="width: 100%; height: 100%">
      <el-table-column prop="name"/>
    </el-table>
  </el-dialog>
  <el-dialog v-model="downloadShowIs" :title="bookInfo.name" width="50%"
             :before-close="()=>{downloadShowIs=false;enableDownloadShowList=[]}">
    <div style="display: flex; flex-wrap: wrap; justify-items: flex-start">
      <el-checkbox
          v-model="downloadVols[i]"
          :label="n"
          size="large"
          v-for="(n,i) in enableDownloadShowList"
          :key="i"
      />
    </div>
    <template #footer>
      <el-button type="primary" @click="downloadBook">
        Download
      </el-button>
    </template>
  </el-dialog>
</template>

<style>
.info_bg {
  width: 100%;
  height: 100%;
  background-color: white;
  padding: 0 !important;

  display: flex;
  flex-direction: column;
}

.info_header {
  width: 100%;
  height: 140px;
  margin: 0 auto;
  color: lightgray;
  background: linear-gradient(to bottom, black, darkgray);

  padding: 20px 0 20px 20px;

  display: inline-flex;
  justify-content: flex-start;
  align-items: center;

  box-sizing: border-box;
}

.info_Des {
  width: 100%;
  height: auto;
  margin: 0 auto;
  padding: 0 20px 20px 20px;
  color: black;

  box-sizing: border-box;

  justify-content: flex-start;
  text-align: left;

  background: linear-gradient(to bottom, darkgray, lightgray);
}

.info_Vol {
  width: 100%;
  margin: 0 auto;
  color: black;

  box-sizing: border-box;

  flex: 1;

  justify-content: flex-start;
  text-align: left;

  background: linear-gradient(to bottom, lightgray, whitesmoke);

  overflow-y: auto;
}

.info_VolItem {
  width: 100%;
  height: 140px;
  margin: 0 auto;
  color: black;
  opacity: 0.8;

  display: inline-flex;
  justify-content: flex-start;
  align-items: center;

  padding: 0 20px;
  box-sizing: border-box;
  border-top: 1px solid dimgray;
  border-bottom: 1px solid dimgray;
  border-left: none;
  border-right: none;
}

.backgroundL {
  width: 100%;
  height: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
}
</style>