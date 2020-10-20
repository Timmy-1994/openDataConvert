using System;
using System.Collections.Generic;
using System.IO;
using System.Text;
using System.Xml;
using System.Xml.Serialization;

namespace OAC_opendata_Console.Model
{

    #region OCM 海象資料


    /// <summary>
    /// 提供 OCM NetCDF 轉檔後存的 Json 格式
    /// </summary>
    public class OcmDataSet
    {
        public List<OcmData> dataset { get; set; }

    }
    public class OcmData
    {
        public OcmHeader header { get; set; }
        public List<object> data { get; set; }
        /// <summary>
        /// 存放本資料集中 data 的最大最小值
        /// </summary>
        public float minimum { get; set; } = 0;
        public float maximum { get; set; } = 0;
        public float average { get; set; } = 0;


    }
    public class OcmHeader
    {
        public int parameterCategory { get; set; }
        public int parameterNumber { get; set; }
        public int scanMode { get; set; }
        public int nx { get; set; }
        public int ny { get; set; }
        public int lo1 { get; set; }
        public int la1 { get; set; }
        public int lo2 { get; set; }
        public int la2 { get; set; }
        //resolution
        public double dx { get; set; }
        public double dy { get; set; }
        public string refTime { get; set; }


    }

    /// <summary>
    /// 下載資料集的日期索引檔
    /// </summary>
    public class OcmDataSetIndex
    {
        public string ApiUrl { get; set; }

        public List<OcmDataIndex> DataList { get; set; }


    }
    /// <summary>
    /// 單一筆資料索引內容
    /// </summary>
    public class OcmDataIndex
    {
        /// <summary>
        /// 索引編號
        /// </summary>
        public int Index { get; set; }
        /// <summary>
        /// 時間集編號
        /// </summary>
        public int Time { get; set; }

        /// <summary>
        /// 資料日期時間
        /// </summary>
        public string timeUTC { get; set; }
        public string time08 { get; set; }



        /// <summary>
        /// nc檔下載來源網址
        /// </summary>
        public string ncDataUrl { set; get; }

        /// <summary>
        /// nc檔存放位址
        /// </summary>
        public string ncFilePath { set; get; }

        /// <summary>
        /// 轉出的 json 檔名
        /// </summary>
        public string name { get; set; }

        /// <summary>
        /// 該筆資料集的最小、最大值
        /// </summary>
        public float min_Intensity { get; set; }
        public float max_Intensity { get; set; }

    }

    // UTC、+8 時區日期物件
    public class DateUtc_08
    {
        public string timeUTC { get; set; }
        public string time08 { get; set; }
        public DateTime dt08 { get; set; }
        public DateTime dtUTC { get; set; }

    }



    #endregion

    #region  運研所海氣象資料轉檔 - 商港觀測資料


    public class isoheStation
    {

        public string Name { get; set; }
        public string Code { get; set; }
        public string XmlSourceUrl { get; set; }
        public string OpendataLinkUrl { get; set; }

    }


    public class StationFieldsDesc
    {
        public string Name { get; set; } = "測站名稱";
        public string Code { get; set; } = "代碼";
        public string XmlSourceUrl { get; set; } = "XML資料來源網址";
        public string OpendataLinkUrl { get; set; } = "OpenData資料集網址";

    }

    public class StationList
    {
        public FieldsDesc FieldsDesc { get; set; }
        public List<Station> Stations { get; set; }

    }

    /// <summary>
    /// 單一類測站的坐標位置
    /// </summary>
    public class StationLocation
    {
        public double Latitude { get; set; } = 0.0;
        public double Longitude { get; set; } = 0.0;
    }

    /// <summary>
    /// 欄位描述
    /// </summary>
    public class FieldsDesc
    {
        public StationFieldsDesc StationFieldsDesc { get; set; }
        public TideDataFieldsDesc TideDataFieldsDesc { get; set; }
        public HistoryDataFieldsDesc HistoryDataFieldsDesc { get; set; }
        public WindDataFieldsDesc WindDataFieldsDesc { get; set; }

    }

    /// <summary>
    /// 單一測站資訊及資料集
    /// </summary>
    public class Station
    {
        public string Name { get; set; }
        public string Code { get; set; } = "代碼";
        public string XmlSourceUrl { get; set; } = "XML資料來源網址";
        public string OpendataLinkUrl { get; set; } = "OpenData資料集網址";
        public StationDataSet DataSet { get; set; }

    }

    public class StationDataSet
    {
        public TideData TideData { get; set; }
        public HistoryData HistoryData { get; set; }
        public WindData WindData { get; set; }

    }


    public class TideDataFieldsDesc
    {
        public string Date_Time { get; set; } = "觀測時間";
        public string TideValue { get; set; } = "潮位(公尺)";

    }
    public class TideDataItem
    {
        public string Date_Time { get; set; }
        public double TideValue { get; set; }
    }
    public class TideData
    {
        public StationLocation location { get; set; }
        public List<TideDataItem> Data { get; set; }

    }

    public class HistoryDataFieldsDesc
    {
        public string Date_Time { get; set; } = "觀測時間";
        public string HS { get; set; } = "波高(公尺)";
        public string TP { get; set; } = "尖峰週期(秒)";
        public string MDIR { get; set; } = "波向(度)";
        public string Tmean { get; set; } = "平均週期(秒)";
        public string Velocity { get; set; } = "流速(公尺/秒)";
        public string Vmdir { get; set; } = "流向(度)";

    }
    public class HistoryDataItem
    {
        public string Date_Time { get; set; }
        public double HS { get; set; }
        public double TP { get; set; }
        public double MDIR { get; set; }
        public double Tmean { get; set; }
        public double? Velocity { get; set; }
        public double Vmdir { get; set; }
    }
    public class HistoryData
    {
        public StationLocation location { get; set; }
        public List<HistoryDataItem> Data { get; set; }

    }

    public class WindDataFieldsDesc
    {
        public string Date_Time { get; set; } = "觀測時間";
        public string WS_AVG { get; set; } = "平均風速(公尺/秒)";
        public string WD_AVG { get; set; } = "平均風向(度)";

    }
    public class WindDataItem
    {
        public string Date_Time { get; set; }
        public double WS_AVG { get; set; }
        public double WD_AVG { get; set; }
    }
    public class WindData
    {
        public StationLocation location { get; set; }
        public List<WindDataItem> Data { get; set; }

    }



    #endregion


    #region 氣象局 OpenData 海面天氣預報 F-A0012-001
    public class DatasetInfo
    {
        public string datasetDescription { get; set; }
        public DateTime issueTime { get; set; }
        public DateTime update { get; set; }
    }

    public class Parameter
    {
        public string parameterName { get; set; }
        public string parameterValue { get; set; }
        public string parameterUnit { get; set; }
    }

    public class Time
    {
        public DateTime startTime { get; set; }
        public DateTime endTime { get; set; }
        public Parameter parameter { get; set; }
    }

    public class WeatherElement
    {
        public string elementName { get; set; }
        public List<Time> time { get; set; }
    }

    public class Location
    {
        public string locationName { get; set; }
        public List<WeatherElement> weatherElement { get; set; }
    }

    public class Dataset
    {
        public DatasetInfo datasetInfo { get; set; }
        public List<Location> location { get; set; }
    }

    public class Cwbopendata
    {
        public string @xmlns { get; set; }
        public string identifier { get; set; }
        public string sender { get; set; }
        public DateTime sent { get; set; }
        public string status { get; set; }
        public string msgType { get; set; }
        public string source { get; set; }
        public string dataid { get; set; }
        public string scope { get; set; }
        public Dataset dataset { get; set; }
    }

    public class F_A0012_001
    {
        public Cwbopendata cwbopendata { get; set; }
    }

    #endregion

   
}


namespace OAC_opendata_Console.Model_F_A0021_001
{

    #region 氣象局 OpenData 海面天氣預報 F-A0021-001

    public class ValidTime
    {
        public DateTime startTime { get; set; }
        public DateTime endTime { get; set; }
    }

    public class WeatherElement
    {
        public string elementName { get; set; }
        public string value { get; set; }
        public object time { get; set; }
    }

    public class Time
    {
        public ValidTime validTime { get; set; }
        public List<WeatherElement> weatherElement { get; set; }
    }

    public class Location
    {
        public string locationName { get; set; }
        public string stationId { get; set; }
        public string latitude { get; set; }
        public string longitude { get; set; }
        public List<Time> time { get; set; }
    }

    public class Dataset
    {
        public List<Location> location { get; set; }
    }

    public class Cwbopendata
    {
        public string @xmlns { get; set; }
        public string identifier { get; set; }
        public string sender { get; set; }
        public DateTime sent { get; set; }
        public string status { get; set; }
        public string msgType { get; set; }
        public string source { get; set; }
        public string dataid { get; set; }
        public string scope { get; set; }
        public string note { get; set; }
        public Dataset dataset { get; set; }
    }

    public class F_A0021_001
    {
        public Cwbopendata cwbopendata { get; set; }
    }

    #endregion

}