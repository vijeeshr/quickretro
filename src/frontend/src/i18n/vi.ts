export default {
  common: {
    anonymous: 'Ẩn danh',
    minutes: 'Phút',
    seconds: 'Giây',
    start: 'Bắt đầu',
    stop: 'Dừng',
    copy: 'Sao chép',
    board: 'Bảng',
    toolTips: {
      darkTheme: 'Bật chế độ tối',
      lightTheme: 'Bật chế độ sáng',
    },
    contentOverloadError: 'Nội dung vượt quá giới hạn cho phép.',
    contentStrippingError:
      'Nội dung vượt quá giới hạn cho phép. Phần văn bản dư đã được cắt bớt ở cuối.',
    invalidColumnSelection: 'Vui lòng chọn ít nhất một cột',
    typing: '{name} đang nhập',
    share: {
      linkCopied: 'Đã sao chép liên kết!',
      linkCopyError: 'Không thể sao chép. Vui lòng sao chép thủ công.',
    },
    customColumnSetup: {
      shareLabel: 'Chia sẻ cấu hình cột',
      shareHelp:
        'Sao chép liên kết bên dưới để chia sẻ cấu hình cột của bạn với người khác, hoặc lưu lại để dùng sau.',
      applied: 'Đã áp dụng cấu hình cột',
    },
  },
  join: {
    label: 'Tham gia với tư cách khách',
    namePlaceholder: 'Nhập tên của bạn tại đây!',
    nameRequired: 'Vui lòng nhập tên của bạn',
    button: 'Tham gia',
  },
  createBoard: {
    label: 'Tạo bảng',
    namePlaceholder: 'Nhập tên bảng tại đây!',
    nameRequired: 'Vui lòng nhập tên bảng',
    teamNamePlaceholder: 'Nhập tên nhóm tại đây!',
    button: 'Tạo',
    buttonProgress: 'Đang tạo..',
    captchaInfo: 'Vui lòng hoàn tất CAPTCHA để tiếp tục',
    boardCreationError: 'Có lỗi xảy ra khi tạo bảng',
    columns: 'Cột',
  },
  dashboard: {
    timer: {
      oneMinuteLeft: 'Còn 1 phút nữa',
      timeCompleted: 'Hết giờ rồi!',
      title: 'Bắt đầu/Dừng bộ đếm',
      helpTip:
        'Điều chỉnh phút và giây bằng nút + và -, hoặc phím mũi tên Lên và Xuống trên bàn phím. Thời gian tối đa là 1 giờ.',
      invalid: 'Vui lòng nhập giá trị phút/giây hợp lệ. Phạm vi cho phép là từ 1 giây đến 60 phút.',
      tooltip: 'Bộ đếm ngược',
    },
    share: {
      title: 'Sao chép và chia sẻ liên kết bên dưới cho người tham gia',
      toolTip: 'Chia sẻ bảng với người khác',
    },
    mask: {
      maskTooltip: 'Ẩn nội dung',
      unmaskTooltip: 'Hiện nội dung',
    },
    lock: {
      lockTooltip: 'Khóa bảng',
      unlockTooltip: 'Mở khóa bảng',
      message: 'Không thể thêm hoặc cập nhật. Bảng đã bị khóa bởi người tạo.',
      discardChanges: 'Bảng đã bị khóa! Các nội dung chưa lưu đã bị hủy',
    },
    spotlight: {
      noCardsToFocus: 'Không có thẻ nào để tập trung',
      tooltip: 'Tập trung vào thẻ',
    },
    print: {
      tooltip: 'In',
    },
    language: {
      tooltip: 'Đổi ngôn ngữ',
    },
    delete: {
      title: 'Xác nhận xóa',
      text: 'Dữ liệu sẽ không thể khôi phục sau khi xóa. Bạn có chắc chắn muốn tiếp tục không?',
      tooltip: 'Xóa bảng này',
      continueDelete: 'Có',
      cancelDelete: 'Không',
    },
    columns: {
      col01: 'Điều làm tốt',
      col02: 'Khó khăn',
      col03: 'Hành động tiếp theo',
      col04: 'Ghi nhận',
      col05: 'Cần cải thiện',
      cannotDisable: 'Không thể tắt cột đang có thẻ',
      update: 'Cập nhật',
      discardNewMessages: 'Bản nháp của bạn đã bị hủy vì cột đã bị tắt.',
    },
    printFooter: 'Được tạo với',
    offline: 'Có vẻ như bạn đang ngoại tuyến.',
    notExists: 'Bảng này đã bị xóa tự động hoặc đã bị người tạo xóa thủ công.',
    autoDeleteScheduleBase: 'Bảng này sẽ được tự động dọn dẹp vào ngày {date}',
    autoDeleteScheduleAddon: ', nên bạn không cần phải xóa thủ công.',
  },
}
