import '../styles/WashMachine.css'

const WashMachine = ({ status, number }) => ( // free, busy, my
    <div className={`washmachine_body ${status}`}>
        <div className='washmachine_buttons'>. . .</div>
        <div className='washmachine_panel'>{number}</div>
        <div className={`washmachine_drum spinning ${status}`}></div>
    </div>
);
export default WashMachine;