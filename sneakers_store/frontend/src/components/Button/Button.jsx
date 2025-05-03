import styles from './styles.module.scss';
import React, {useState} from 'react';

export const Button = () => {

  const [isChecked, setIsChecked] = useState(false)
  const imageSrc = isChecked ? "/img/checked.svg" : "/img/plus.svg"

    return(
        <button className={styles.btn} onClick={() => setIsChecked(!isChecked)}>
          <img width={30} height={30} src={imageSrc} alt="plus" />
        </button>
    )
}