import { Container } from "@maxhub/max-ui";
import Header from "../../components/Header";
import Footer from "../../components/Footer";
import { LuUser, LuCheck, LuX, LuUsers, LuDoorOpen } from 'react-icons/lu';
import '../../styles/Room.css';

const RoomResident = () => {

    //  TODO Тестовые данные, потом брать с json
    const roomData = {
        number: '404',
        floor: 4,
        status: 'normal',
        residentsCount: 3,
        maxResidents: 3,
        userStatus: 'Студент',
        cleanDays: {
            currentDay: 23,
            goalDay: 30
        }
    };

    const getStatusInfo = (status) => {
        switch (status) {
            case 'normal':
                return {
                    class: 'cool',
                    text: 'Нет замечаний',
                    icon: <LuCheck size={16} />
                };
            case 'bad':
                return {
                    class: 'bad',
                    text: 'Есть замечания',
                    icon: <LuX size={16} />
                };
            default:
                return {
                    class: 'cool',
                    text: 'Нет замечаний',
                    icon: <LuCheck size={16} />
                };
        }
    };

    const statusInfo = getStatusInfo(roomData.status);

    return (
        <div style={{ minHeight: '100vh', display: 'flex', flexDirection: 'column' }}>
            <Header />
            <Container>
                <div className="room-content">
                    <div className="room_status">
                        <div className="room_status_info">
                            <h4>
                                Комната {roomData.number}
                            </h4>
                            <h5 type="secondary">
                                Этаж {roomData.floor}
                            </h5>
                        </div>
                        <div className={`room_status_badge ${statusInfo.class}`}>
                            {statusInfo.icon}
                            {statusInfo.text}
                        </div>
                    </div>

                    <div className="room_stats">
                        <div className="room_stat_card">
                            <LuUsers size={24} />
                            <div>
                                <div className="stat_number">{roomData.residentsCount}</div>
                                <div className="stat_label">Жильцов</div>
                            </div>
                        </div>
                        <div className="room_stat_card">
                            <LuDoorOpen size={24} />
                            <div>
                                <div className="stat_number">{roomData.maxResidents}</div>
                                <div className="stat_label">Мест</div>
                            </div>
                        </div>
                        <div className="room_stat_card">
                            <LuUser size={24} />
                            <div>
                                <div className="stat_number">{roomData.userStatus}</div>
                                <div className="stat_label">Ваш статус</div>
                            </div>
                        </div>
                    </div>

                    <div className="room_clean_progress">
                        <span>Прогресс снятия с санитарного контроля</span>
                        <div className="progress_wrapper">
                            <progress
                                className="progress_bar"
                                value={Math.min((roomData.cleanDays.currentDay / roomData.cleanDays.goalDay) * 100, 100)}
                                max={100}
                            />
                            <span className="progress_label">
                                {Math.min(Math.round((roomData.cleanDays.currentDay / roomData.cleanDays.goalDay) * 100), 100)}%
                            </span>
                        </div>
                    </div>
                </div>
            </Container>
            <Footer />
        </div>
    );
};

export default RoomResident;